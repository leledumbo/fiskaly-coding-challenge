# How to run the project and tests

In any cases, ensure you have Go 1.20+ accessible from PATH

## Run

1. Open any terminal then navigate to this README's folder, then to signing-service-challenge-go,
   where you can see main.go as the entry point file
2. Run `go mod tidy`
3. Build the executable with `go build`
4. Run the resulting executable `signing-service-challenge-go`:
   Windows: `>signing-service-challenge-go`
   Linux/macOS: `$ ./signing-service-challenge-go`
5. The executable will listen to port 8080 locally

## Endpoints you can hit

The HTTP client assumed here is curl, adjust accordingly if you use a different one

1. create device signature

   `curl localhost:8080/api/v0/create_device_signature -d '{"device_id":"a","algorithm":"rsa"}'`

2. sign transaction

   `curl localhost:8080/api/v0/sign_transaction -d '{"device_id":"a","data":"some data"}'`

2. verify signature

   `curl localhost:8080/api/v0/verify_signature -d '{"device_id":"a","data":"<signed data returned by sign transaction>","signature": "<signature returned by sign transaction>"}'`

3. list devices

   `curl localhost:8080/api/v0/list_devices`

## Test

1. Open any terminal then navigate to this folder
2. Run `go test ./...` for built-in test to work
3. For convenience, the underlying test framework provides a realtime web UI, this can be done by:

   `go install github.com/smartystreets/goconvey`

   followed by:

   `$GOPATH/bin/goconvey`

   certainly you must have GOPATH environment variable set up right.
4. Additionally, a test coverage report can be generated and viewed in the browser by executing:

   `./make-test-coverage-report.sh`

   sorry, Linux/macOS only, but you can copy the command inside (minus the header) in Windows cmd/PowerShell

# Design decisions and trade-offs

I'll write this in story style so you can see my POV when working on the challenge. If you want a
straightforward version, feel free to jump to [summary](#summary).

The first thing I do is understanding the directory structure and see what has been provided for me
as a starting point. I see no use of 3rd party framework, only Go's standard "net/http" with
"encoding/json" so that saves a lot of time for me as my experience with Go's 3rd party framework
is not much. My longest work with it was actually writing our own framework, for the sake of
throwing away dependencies on the work of some people we can't push when we have problems. I have
only worked with one framework: Beego, and that's even a rather old version (1.x), where it changed
in 2.x that made us cannot migrate due to our own modification of the internal.

Back to the challenge, so the only endpoint I'm provided with is /health, that does nothing but
respond HTTP 200 with "pass" body if the method is GET, and HTTP 405 otherwise. It is implemented
in health.go and despite sharing the api package, I don't like how it is registered in server.go.
If endpoint modularization is to be expected, then better make everything modular. In this case, I
separate the routes into its own package and move the endpoint registration to route file. To aid
that a RegisterRoute function is introduced in server.go meant to be called by endpoint file's
init() function. \*http.ServeMux instance creation in Run() method is refactored as a lazily loaded
singleton function. This way, adding or removing endpoints is as simple as adding or removing
relevant endpoint files, no change to server.go is required. Health() function also has its bound
to Server removed, because not only Server has nothing for the endpoint to access (except
listenAddress, but also not useful), but also neither of the auxiliary functions (WriteAPIResponse,
WriteErrorResponse, WriteInternalError) is linked to Server, so they can be called standalone.
(approximate time spent undisturbed: 15-30 mins)

Next I'm examining the crypto package. The situation is quite similar to api package, everything in
generation.go should belong to each algorithm's file. Despite both implementations are named
XXXMarshaler, for some obvious reasons ECCMarshaler uses (En|De)code instead of [Un]Marshal. To
easily extend with new algorithms, similar to easily adding a new API endpoint, interfaces will
be created to hold method headers each algorithm should implement with consistent signature, as
well as a function to register the algorithm. Again, the API endpoint that will select and utilize
them will have no idea how many algorithms are actually registered, though it can know whether a
requested algorithm exists or not and will respond appropriately.
(approximate time spent undisturbed: 15-30 mins)

I decide to focus on one algorithm implementation first, as if that one is already working, adding
more is easy peasy. The first endpoint to implement is create signature device, which leads me to 
examine domain and persistent package as the storage model and layer. The first one contains
practically nothing, so I'm free to design. I first use another interface-implementation approach,
but after some careful consideration, there is no need for interface. Any code that depends on it
should use it as is, including persistence package. I remember README contains this line:
"Encoding / decoding of different key types (only needed to serialize keys to a persistent storage)"
as well as:
"For now it is enough to store signature devices in memory. Efficiency is not a priority for this.
In the future we might want to scale out. As you design your storage logic, keep in mind that we may
later want to switch to a relational database."
So I'm thinking the model has to be something that can be easily persisted, cannot contain an
interface. Reading RSAMarshaler.RSAMarshaler I discover that you can generate both public and
private key only from private key, so I decide there's no need to persist the public key to save
some space.
(approximate time spent undisturbed: 30-60 mins)

I change interfaces and struct fields a couple of times before getting the right one, as this
endpoint implementation touches all packages. As can be seen from the implementation, the code is
free from specific algorithm or key pair, the only concrete object is device, which is also free
from any specific persistence implementation
(approximate time spent undisturbed: 60-90 mins)

The second endpoint to implement is actually to verify the first endpoint implementation: list
devices. I only add List() to the persistence interface for this, the endpoint code only wraps
that lightly.
(approximate time spent undisturbed: 5 mins)

The third endpoint to implement is sign transaction. I need a little reading on the subject as my
crypto knowledge isn't that vast I almost mistaken sign for encrypt, thankfully crypto package
documentation is enough to enlighten me. I got stuck here wondering how to generate the key pair
again in an abstract manner, as what I store is only the PEM encoded version of the private key.
Eventually I introduce ConstructKeyPair() that calls Deserialize() on the correct KeyPair. This
allows me to call Sign() method of the chosen algorithm safely without leaking abstraction.
(approximate time spent undisturbed: 90-120 mins)

The final endpoint is a bonus: verify signature. What's the point of signing if you can't verify?
Thankfully this one is also quite straightforward, almost 1:1 copy of sign with difference in the
last call. With this implemented, KeyPair.PublicKey() finally has its usecase.
(approximate time spent undisturbed: 15 mins)

I realize I haven't fulfilled these constraints:

* The system will be used by many concurrent clients accessing the same resources.
* The signature_counter has to be strictly monotonically increasing and ideally without any gaps.

Usually such a functionality is delegated to something like Redis or relational database ability
to ensure and guarantee atomic operation, but as it is expected for the underlying storage to
change, we cannot rely on that. Thus, this must be implemented at application level. I'm already
used to sync package especially WaitGroup and Mutex for concurrency management, but blocking the
whole sign transaction doesn't sound wise. Asking ChatGPT, I'm pointed towards sync.Map, where it
is possible to lock object usage of the same id while letting different ones to run concurrently.
To keep the storage abstraction intact, this atomic layer will be implemented on top of Storage
interface.
(approximate time spent undisturbed: 30 mins)

Finally, testing. From experience, I've found that go convey adds nice BDD style readings over Go's
built-in testing capabilities. You can literally read in human language what each test is doing in
tree-like style. Adding GoMock to the recipe makes things even better due to its capability to
expect how many times a function will be called with what arguments and what it will return, a
runtime patch to the actual implementation. I ask ChatGPT to generate the test as otherwise it's
going to be the most time consuming task of all. It's easy, but repetitions of the boilerplate
wastes a lot of time. This is one usecase I will advice anyone to make a good use of AI for. I
still need to add some cases it fails to "see", though.
(approximate time spent undisturbed: 90-120 mins)

## Summary

* Modularization of endpoints and crypto algorithms means that you no longer have a single file or
  entry point to see all available options in one place, which itself is potentially an abstraction
  leak, with ease of addition/removal through a single file in return
* Crypto algorithms are also modularized, every implementation should only reside in a single file
* Crypto layer interface uses any for Key as there is no way to generalize public/private key used
  by different algorithms but it should still be well protected from the other interface properties
* Create signature device endpoint optionally allows update, to ease changing of algorithm and/or
  label, in which due to this signature counter, and therefore last signature, will naturally reset
* ID exists both in Device definition and as key to save/load for ease of use with double amount
  of memory used as cost (the compiler can optimize this away, by using the same string reference,
  but don't expect this to be a guarantee)
* For completeness sake, I also provide endpoints to verify the signature, as it seems incomplete
  if you can sign it without verifying

# Any assumptions and known limitations

* I assume all possible crypto algorithms are always based on public/private key pair
* README mentions "Library to generate UUIDs, included in go.mod" but go.mod only contains this
  module, no other dependencies mentioned, so I assume it is not required, as from the task
  description, it looks like so
* CreateSignatureDeviceResponse contains nothing, as all data required by the client to do
  further operations, mainly device ID, is already on their hands. The space is still provided in
  case someday something needs to be returned for whatever reason
* Signature counter is kept per device instead of per call to sign transaction
* Signature counter and last signature will only be updated on a successful signing attempt
* I don't quite understand "`list / retrieval operations` for the resources generated in the
  previous operations should be made available to the customers." and given the time constraint as
  as well as the days to work on it (weekends, basically), I don't think I can ask about it. So I'm
  assuming a list of devices instead of every signature and signed data because it doesn't sound
  sensible to me to keep all of them (should be easy to add, though, it's just audit trail or logging)
* crypto interface contains an alias type Key for any, which is quite difficult to debug if you
  implement a new algorithm by copy pasting from existing one, as there is no compile time
  protection for GenerateKeyPair/ConstructKeyPair if you accidentally return the wrong actual
  KeyPair implementation (trust me, I experience this myself)
* As commonly done in a monolithic environment, the same single persistence layer implementation
  will be used throughout the application. Therefore, there is no multi implementation support
  available. You can change the selected implementation inside persistence/instance.go init()
  function, where the instance is initialized

# Use of AI

I only use ChatGPT throughout the challenge, mainly for asking:

* How to deal with github related issue since I use private repo to upload this (obviously I don't want
  anyone else to copy my solution)
* How to avoid cyclic dependencies when modularizing the endpoints since the routes
  require a way to access ServeMux, but Server also imports the routes, which ended up ServeMux placed
  in its own package and imported by both server and routes
* How to implement custom UnmarshalJSON so that I can validate alongside parsing instead of after,
  decluttering the endpoint implementation
* How to use encoding/base64 and the various crypto algorithms (implicity sha256) to sign and verify
* What PEM is (I've used it, but I know neither the expansion nor the meaning)
* Safest approach to abstract public and private key
* How to implement atomic increment over concurrent HTTP requests

# Approximate time spent

Summing approximate time spent from [design decisions and trade-offs](#design-decisions-and-trade-offs)
section:

* fastest: 15 + 15 + 30 + 60 + 5 +  90 + 15 + 30 + 90 = 350 minutes
* slowest: 30 + 30 + 60 + 90 + 5 + 120 + 15 + 30 + 120 = 500 minutes

This README alone also takes quite significant amount of time, I would say perhaps 60 minutes in total.
That brings the approximate time spent to 410-560 minutes or about 6:50-9:20. I'm pretty sure this is
quite accurate counting resting, eating, bathing, and house chores time, leaning more towards the slowest.
The first two were done on Friday evening, as I had another gig to do after the interview, while the rest
but the last was fully done on Saturday, starting at around 10AM to 10PM. Sunday is full only to implement
the testing.
