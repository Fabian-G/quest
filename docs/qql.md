# Quest Query Language (QQL)

## Syntax and Semantics

### Durations

## Available Functions

## Macros

## A Word on Performance

Since QQL is effectively a model checker for first-order logic (which is a PSPACE-Complete Problem) and 
since this implementation is certainly not the breakthrough we have been waiting for, the performance 
of running your query may theoretically be extremely bad. 
But unless you are dealing with a massive todo file or nest your quantifiers miles deep you will probably be fine. 
If for whatever reason you are facing performance issues anyway here are a few things you can do:

1. Use `quest archive`. This will move your completed tasks to a seperate file and will therefore reduce the number of tasks that need to be checked against your query.
2. Make use of **short-circuiting** by moving the expensive operations (e.g. quantification over large collections) to the end of your query. For example `!done(it) && expensiveMacro(it)` could be significantly faster than `expensiveMacro(it) && !done(it)`, because in the former case `expensiveMacro(it)` will only be checked on tasks that are not done.
3. Do not use `shell` or `command` functions if at all possible. These operations are not only asymptotically slow, but are actually slow in practice. So if you can, avoid using them. Otherwise refer to point 2.