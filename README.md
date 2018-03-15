
The job of any application is to _do things_ (To put it simply). Sometimes, those things may take time, and the application cannot wait to make sure they are done. But still, it's nessecary that they are.

An example of this kind of problem that we face at GO-JEK is when a new user signs on to our platform. When this happens, we need to create multiple profiles for the user on our different services, and at the same time give the user instant confirmation that thier signup was successful. 

[[user profile creation]]

A few problems arise if we treat this as a synchronous problem:
- How can we give the newly signed up user instant confirmation of their signup, if their profile creation on our system takes more time?
- If we decide to give the user an optimistic confirmation (which makes the confirmation instant, but assumes that all profile creations will be successful), then how can we _ensure_ that all profile creations are, indeed, successful?

## Jobs, and batch processing

Whenever there is a need to do a bunch of tasks after the occurence of some event (in our example, this event refers to a new user signing up), we can treat each task as a "job". A job is not part of the the response to the event that took place (the response, in our example was the instant confirmation that we gave to our user of their successful signup), and takes place _asynchronously_. It's important for us to classify jobs based on their importance:
- __Critical jobs__ are those that cannot afford to fail. In our case, if a new user signs up, we cannot afford to skip ther profile creation in any of our services. Another example would be sending the user a notification when their GO-PAY balance has changed.
- __Non-critical jobs__ like sending user signup statistics to our analytics service, or sending promotional notifications to a user, can afford to fail (of course, we strongly prefer that the don't 😅)

In any system that we design, non-critical jobs _should not fail_, and critical jobs _cannot fail_.

[[system with jobs in it]]

One simple mechanism to make sure that jobs don't fail is to just [retry](https://blog.gojekengineering.com/how-go-jek-handles-microservices-communication-at-scale-5ad91be98c77#d671) them when they do. This may work in smaller systems, but problems arise once you start to scale:
- If too many jobs fail on a single system, they start to build up, and after some time your system can run out of resources
- If the machine itself fails, then _all_ jobs that were running, or supposed to run on it fail as well.

[[failing systems]]

Both of these scenarios are unacceptable.

## Job queues

The solution that we use to solve most of our batch processing woes is to set up a worker queue. The way we do this is to have a description of the job itself pushed onto a queue that resides on a different process from the job creator (or publisher). We then create workers that subscribe to the job queue and execute each job in sequence.

[[job queue diagram]]

Implementing this kind of a job queue solves a lot of problems for us:

### Resiliency

Once we have a job queue running, we protect ourselves from job failures as well as application failures. Messaging queues like RabbitMQ have an acknowledgement mechanism, which clears a job from a queue _only_ once it's acknowledged (or optionally requeued when negatively ackowledged). 

Job failures are handled by negatively acknowledging the job, which will then be requeued and handled by another worker.

If the worker itself fails while processing jobs from a queue, those jobs will simply not be acknowledged at all, and will be redistributed if the connection between the worker and the message queue closes.

You can make the jobs themselves resilient by using redundancy among messaging queues. This is implemented in frameworks like Apache kafka. This ensures that even if the job queue itslef fails, it won't lead to the jobs not being done.

### Scalability

By moving the workload to another process, it's now possible to scale out our workload. If we find that there are too many jobs and they're not getting done in time, all we have to do is increase the number of worker processes to accomodate the workload.

We can do the same thing with the job publishers, if we find that the number of requests are more than we can accomodate.

### Load distribution

By separating the place at which event creation requests are received, and the place at which they are processed, we have essentially distributed the load such that an increase in load at one end will not affect the system at the other end. 

In our case, this means that if there is some delay that affects profile creation on our services, that delay will not affect the response time and the experience of the user signing up. We make these two processes mutually exclusive, as they should be.

