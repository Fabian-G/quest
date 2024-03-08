# Quest Score

> Note: The Quest Score is more or less experimental 
> in the sense that it is likely to change in the future.

The Quest Score aims to provide some general guideline on which task to do next.
It is loosely based on the concept of the Eisenhower Matrix, which sorts 
a task in a two dimensional matrix. The first dimension is the urgency and
the second dimension the importance.

The Quest Score measures urgency in terms of a configurable date (usually the date associated 
with the due tag) and the closer this date gets the higher is the urgency.

The importance on the other hand is measured by the priority of a task.
A higher priority (closer to A) means the task is more important.

After that Quest will take this two dimensional vector and try to calculate 
a scalar value between 0 and 10 from this in a meaningful way. 
Currently it does so by calculating the squared mean.

The quest score comes with a sensible default [configuration](configuration.md). So to use it, just put
`score` inside some of your projections.

## Make old task urgent

You might want to let tasks gain slightly in urgency the older they get too avoid 
having them lying around forever.
This is what the `urgency-default` option in the `[quest-score]` section is for.
If the tasks does not have any of the configured `urgency-tags` and the `urgency-default`
is set to a non-zero duration value, Quest will consider the creation date plus the 
`urgency-default` as the urgency date.
This has one minor drawback though. Imagine you know 6 months in advance that you need to bake a cake.
Assuming you use the t tag according to the convention, you would probably do:
```bash
quest add bake a cake t:6m
```
This new task will have the creation date of today, so when it shows up in 6 months it will
have a very high Quest Score, which is not what we want.

To fix this we can add the t tag to the list of urgency tags, but offset by some amount of time.

```toml
[quest-score]
urgency-tags = [ "due", "t+2m" ]
urgency-default = "2m"
```

Now Quest will first consider the due tag and if that is not set it will consider the t tag offset by two months.
So in our example the "bake a cake" task will have maximal urgency 2 months after it shows up.

