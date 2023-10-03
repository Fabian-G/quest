# Quest Query Language (QQL)

Quest comes with a powerful query language, which is based on first-order logic (FOL). 
If you are already familiar with FOL you should be able to get started in no time.
If not, there may be a learning curve if you want to write more advanced queries, but the basics should be pretty straightforward.

A basic QQL query will look like this:
```
!done(it) && priority(it) >= prioC && substring(description(it), "foo")
```
Quest will run your query against each item in the list and checks if it matches. 
The query will receive the current item through the `it` variable.
Therefore the above query matches all items that are not done, have a priority higher or equal to (C) and contain the word "foo" in the description.

## Syntax and Semantics

The syntax will not be explained in great detail here, because it is basically what you would expect from any programming language.

A primary expression can either be a function call of the form `function(args)` (see [Available Function](#functions)), 
an integer value, a boolean value (`true` or `false`), a string literal (with double-quotes `"hello world"`), a duration literal (see [Durations](#durations)) or a constant (see [Constant](#constants)).

There are the following boolean operators: `!` (not), `&&` (and), `||` (or), `->` (implication, "if a then b" and equivalent to "!a || b"). 
You can always use parentheses, but if you don't, the order they are listed here is the order of precedence.

Example:
```java
true -> false || true && !false // evaluates to true
```

To compare values you can use the usual `<`, `<=`, `==`, `>=` and `>` operators. 
You can compare dates, integers, strings, durations, boolean values and tasks, but the last two can only be compared with `==`.
If you want to check for inequality you can use `!(a == b)`.

Example:
```java
ymd(2022, 2, 2) >= ymd(2021,2,2) && 5 < 6 && "foo" > "bar" // evaluates to true
```

QQL also supports the numeric operators `+` and `-`. 
Obviously this works on the int type (`5+5==10` evaluates to true), but you can also use this for dates and durations (`ymd(2022,2,2)+5d==ymd(2022,2,7)`).

Finally you can use the quantifiers `exists` and `forall` over any collection.
The basic syntax is: "*quantifier* x in *collection*: *expression*". Where *quantifier* is either exists or forall, x is an arbitrary variable name and collection is any collection.
`exists x in collection: expression` will evaluate to true if there is at least one item in the collection so that the expression evaluates to true (with the variable `x` set to that particular item).
Similarly `forall x in collection: expression` evaluates to true if `expression` is true for all items in the collection.
There are multiple ways to obtain a collection to quantify over:

1. Use the `items` constant, which contains all items in your list: `forall item in items: done(item)` (evaluates to true if all items in the list are done)
2. Use `projects` or `contexts` function: `exists proj in projects(it): proj == "+foo"` (evaluates to true if `it` is in the project "+foo")
3. Use `list` to transform any string (for example from a tag) to a list: `exists num in list(tag(it, "favorite-numbers")): num == "2"` (matches for example `favorite-numbers:1,2,3`)

A common pattern is to make a statement about all items that fulfil a certain precondition. This is where the implication operator comes in handy. 
For example if we wanted to check if `it` is the last not done item in a group of items we could write: `!done(it) && forall other in items: tag(it, "group") == tag(other, "group") && !(it == other) -> done(other)` "match it if: it is not done and all other items that have the same group-tag as it are done".

### Constants

### Durations

### Functions


## Macros

## A Word on Performance

Since QQL is effectively a model checker for first-order logic (which is a PSPACE-Complete Problem) and 
since this implementation is certainly not the breakthrough we have been waiting for, the performance 
of running your query may theoretically be extremely bad. 
But unless you are dealing with a massive todo file or nest your quantifiers miles deep you will probably be fine. 
If for whatever reason you are facing performance issues anyway here are a few things you can do:

1. Use `quest archive`. This will move your completed tasks to a separate file and will therefore reduce the number of tasks that need to be checked against your query.
2. Make use of **short-circuiting** by moving the expensive operations (e.g. quantification over large collections) to the end of your query. For example `!done(it) && expensiveMacro(it)` could be significantly faster than `expensiveMacro(it) && !done(it)`, because in the former case `expensiveMacro(it)` will only be checked on tasks that are not done.
3. Do not use `shell` or `command` functions if at all possible. These operations are not only asymptotically slow, but are actually slow in practice. So if you can, avoid using them. Otherwise refer to point 2.