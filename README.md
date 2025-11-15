# fund78

Working on trading systems, a property not always empahized in other software projects is determinism.

Using a funnel, such as a queue, which takes in each piece of work one by one allows this to occurr. In essence, make the system single threaded when it cannot be wrong.

Companies have built on this idea, such as the LMAX disruptor. My idea with this project is to figure out how far one can go with just a queue.

Using Rust should help naturally keep the performance high, without the need to worry about GC collections.

The initial goal of mine is to see how many requests per second this application can handle given a consistent stream of requests that take less than 5 ms on average to complete.

The perceived latency for any request from a user should always be less than 50 ms.

Therefore, an initial goal is for the following:

   * Max sustained RPS = 200 RPS (17.8 million requests per day)
   * Max latency-safe burst (accepted in 1s) ≈ 209 RPS
   * Max requests fully completed within any 50 ms user-perceived window = 10 requests (i.e. still 200 RPS equivalent)
