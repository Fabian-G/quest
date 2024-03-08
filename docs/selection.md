# Selecting Tasks

When selecting tasks you can decide between 3 options: Range Query, String Search or Quest Query Language (QQL).
QQL can do everything the other two can do (and more), but might be slightly less convenient.
You can specify which type of query your are using on the command line with the `-q`, `-r` and `-w` options. 
If you don't and simply pass your query as an argument, quest will try to guess what type of query you are using.
"Guessing" means in this case that it tries to parse it as a QQL query, if that fails it tries to parse it as a range query and if that fails it finally treats it like a string search.
While this is convenient most of the time, this can behave unexpectedly if you are intending to write a QQL query, but have a syntax error, which is then silently ignored. 
In that case you should be explicit about the query type and use `-q`.


## Range Query

You can select items by their line number using simple range expressions like:

- `1,2,3`: selects tasks 1, 2 and 3
- `1-3`: Also selects tasks 1, 2 and 3
- `1,3-`: Selects tasks 1 and all tasks after and including task 3

## String search

The string search (usually `-w` flag in the CLI) will do a simple case-insensitive substring search in the task description.

## Quest Query Language (QQL)

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

### Syntax and Semantics

The syntax will not be explained in great detail here, because it is basically what you would expect from any programming language.

A primary expression can either be a function call of the form `function(args)` (see [Available Function](#functions)), 
an integer value, a boolean value (`true` or `false`), a string literal (with double-quotes `"hello world"`), a duration literal (see [Durations](#durations)) or a constant (see [Constant](#constants)).

There are the following boolean operators: `!` (not), `&&` (and), `||` (or), `->` (implication, "if a then b" and equivalent to "!a || b"). 
You can always use parentheses, but if you don't, the order they are listed here is the order of precedence.

Example:
```java
true -> false || true && !false // equivalent to (true -> (false || (true && (!false))))
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
The basic syntax is: `quantifier x in collection: expression`. Where *quantifier* is either exists or forall, x is an arbitrary variable name and collection is any collection.
`exists x in collection: expression` will evaluate to true if there is at least one item in the collection so that the expression evaluates to true (with the variable `x` set to that particular item).
Similarly `forall x in collection: expression` evaluates to true if `expression` is true for all items in the collection.
There are multiple ways to obtain a collection to quantify over:

1. Use the `items` constant, which contains all items in your list: `forall item in items: done(item)` (evaluates to true if all items in the list are done)
2. Use `projects` or `contexts` function: `exists proj in projects(it): proj == "+foo"` (evaluates to true if `it` is in the project "+foo")
3. Use `list` to transform any string (for example from a tag) to a list: `exists num in list(tag(it, "favorite-numbers")): num == "2"` (matches for example `favorite-numbers:1,2,3`)

A common pattern is to make a statement about all items that fulfil a certain precondition. This is where the implication operator comes in handy. 
For example if we wanted to check if `it` is the last not done item in a group of items we could write: `!done(it) && forall other in items: tag(it, "group") == tag(other, "group") && !(it == other) -> done(other)` "match it if: it is not done and all other items that have the same group-tag as it are done".

#### Syntactic Sugar

Function calls come with two rules that make the average usage of QQL a little more convenient:

1. If a function call requires an item as an argument and it is omitted, it will simply be replaced by *it*. For example writing `done()` is the same as `done(it)`
2. If the argument list is empty, the parentheses can be omitted: `done()` is the same as writing `done`

For example: `!done && priority >= prioC` matches all tasks that are not done and have a priority of at least (C).

Another extension to the syntax is that you can check if *it* is in a specific (sub) project by using the project matcher syntax: `+foo`.
The same is possible for contexts with `@foo`.

Example: `@foo && +bar` matches all items that are in a (sub) context of foo (e.g. @foo, @foo.bar, ...) and in a (sub) project of bar.

Note that the project matcher syntax introduces a slight ambiguity between the numeric operation `+` and the project matching.
To make sure that a `+` is interpreted as the numeric operation make sure to put a space behind it (e.g. `5+ aNumber`).

#### Constants

The following constants are available:

| Constant | Description |
| --- | --- |
| it | The task item that the query is currently evaluated against |
| items | The list of all task items |
| maxInt | The maximum integer value |
| minInt | The minimum integer value |
| today | todays date |
| maxDate | The maximum date value |
| minDate | The minimum date value |
| prio* | The priority *. Where * is a letter between A and Z. |
| prioNone | The priority that signals the absence of a priority |

### Durations

A duration literal follows the syntax `span unit`, where *span* is a (possibly negative) integer value and unit one of:

- days (or d) 
- weeks (or w)
- months (or m)
- years (or y)

Examples: `+5y`, `-5days`, ...

#### Functions

A short explanation of the notation: If in the following table the function definition says for example `func(a: int, b: date = minDate): bool`, 
this means that the function with the name `func` takes two arguments. The first one must be of type `int` and the second of type `date`.
The second argument is optional and can be omitted. If it is omitted it will be set to `minDate`. The return type of the function is `bool`.

| Function | Description |
| --- | --- |
| line(i: item): int | The line number of i |
| done(i: item): bool | Whether or not i is already completed |
| description(i: item): string | The description of i (including all tags, projects and contexts) |
| creation(i: item, default: date = minDate): date | The creation date of i if it is set or default otherwise |
| completion(i: item, default: date = maxDate): date | The completion date of i if set or default otherwise |
| projects(i: item): []string | The projects that are present in the description of i in a list. The elements of the list do contain the leading + symbol |
| contexts(i: item): []string | The contexts that are present in the description of i in a list. The elements of the list do contain the leading @ symbol |
| priority(i: item): priority | The priority of i. If no priority is set on i `prioNone` is returned |
| dotPrefix(s: string, prefix: string): bool | Checks if *s* starts with all the dot delimited segments of *prefix*. This is useful if you want to use sub projects or contexts. Examples: `dotPrefix("+foo.bar.baz", "+foo.bar") == true`, `dotPrefix("+foo.bar.baz", "+foo.b") == false` |
| substring(s: string, sub: string): bool | Tests if *s* contains substring *sub* |
| ymd(year: int, month: int, day: int): date | Constructs a date from the provided year, month and day |
| date(yyyymmdd: string, default: date = minDate): date | Parses the given argument into a date (format YYYY-MM-dd). If the format does not match the default is returned |
| tag(i: item, key: string, default: string = ""): string | Returns the value of the first occurrence of the tag with key *key*. If *key* is not set *default* is returned |
| list(l: string): []string | Splits the value l at ",". For example `list("1,2,3")` becomes the list with the elements 1,2 and 3. |
| int(num: string, default: int = 0): int | Parses *num* as an integer. If *num* is not a valid integer *default* is returned | 
| shell(i: item, cmd: string): string | Runs *cmd* using bash. See ([shell and command](#shell-and-command)) |
| command(i: item, cmd: string): string | Same as shell, but runs the cmd directly without bash |

##### Shell and Command

If you want to write really exotic queries you can resort to shell and command, but be aware that using these functions comes with a high performance penalty.
The way this works is that the command will receive a json representation of the specified task (like the one you get with `--json`) on stdin.
Whatever is output on stdout will then be space trimmed and returned.

Example (Match all tasks that were created during full moon):
```
int(shell("jq -r .creation\ \|\ sub\(\"-\"\;\"\"\;\"g\"\)+\"00\" | xargs pom | sed s/\[^0-9\]//g"), 0) >= 94
```

As you can see in the above example, properly escaping the string is not fun. Therefore I would recommend to always put the command in a separate file and 
then use the `command` function instead.
 
### Macros

Macros are a way to give a name to parts of your query so that your queries can be reused and look cleaner.
A simple macro definition in your `config.toml` might look like this:

```toml
# Use "after" and "id" tags to express dependencies between tasks
[[macro]]
name = "blocked" # How you would like to call your macro
args = ["item"]  # The types of arguments that the macro expects. The parameters will be available as arg0, arg1, ...
result = "bool"  # The return type of this macro
query = '!(tag(arg0, "after", "") == "") && (exists pre in items: tag(pre, "id", "") == tag(arg0, "after") && !done(pre))'
inject-it = true # Whether or not a missing item parameter should default to "it". Because this is true we can just write "blocked" instead of "blocked(it)"
```

Note that if you reference other macros from within your macro definition, the other macros must appear before this definition in your config file.

With this definition in place you can then write queries like: `!done && !blocked` to find not completed tasks that are not blocked.

### A Word on Performance

Since QQL is effectively a model checker for first-order logic (which is a PSPACE-Complete problem) and 
since this implementation is certainly not the breakthrough we have been waiting for, the performance 
of running your query may theoretically be extremely bad. 
But unless you are dealing with a massive todo file or nest your quantifiers miles deep you will probably be fine. 
If for whatever reason you are facing performance issues anyway here are a few things you can do:

1. Use `quest archive`. This will move your completed tasks to a separate file and will therefore reduce the number of tasks that need to be checked against your query.
2. Make use of **short-circuiting** by moving the expensive operations (e.g. quantification over large collections) to the end of your query. For example `!done(it) && expensiveMacro(it)` could be significantly faster than `expensiveMacro(it) && !done(it)`, because in the former case `expensiveMacro(it)` will only be checked on tasks that are not done.
3. Do not use `shell` or `command` functions if at all possible. These operations are not only asymptotically slow, but are actually slow in practice. So if you can, avoid using them. Otherwise refer to point 2.
