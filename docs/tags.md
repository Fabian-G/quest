# Tags

By default Quest does not make any assumptions about what tags you use.
You can, however, choose to tell quest more about that to get some extra benefits
like validation and tag expansion.

For example to configure a tag for the `due` date you would put the following lines in your `config.toml`
```toml
[tags.due]
type = "date"
```

Now quest will not only validate that the the value of `due` is a valid date, but it will also allow you to write things like `quest add Clean the stables due:tomorrow`, which will expand `tomorrow` to the date of tomorrow before writing to your todo.txt.

The available options in a `[tags.my-tag]` section are:

| Option | Description |
| --- | --- |
| type | One of `date`, `int`, `duration` or `string`. This decides which validations apply and which tag expansions can be used. The "string" type does not have validation or any expansion. |
| humanize | When this tag is displayed in a column it will be printed in human friendly format (currently this only has an effect with the date type) |
| styles | A list of style definitions that apply only to this tag (see [Styling](styling.md)) |

## Int Type

When the `int` type is configured the validation will make sure that the value provided with the tag is a valid integer value (positive or negative).
The tag value can also be of the form `base[+-]x` where `x` is a valid integer and `base` is one of `min`,`max`,`pmin`,`pmax`.
`min` (`max`) will be replaced with the minimum (maximum) integer value that is currently present **on that tag** on **any** task. 
`pmin` (`pmax`) will be replaced with the minimum (maximum) integer value that is currently present **on that tag** on any task that **shares** at least one **project**. 

## Date Type

When the `date` tag is configured the validation will make sure that the value provided is a valid date of the format `YYYY-MM-dd`.
Strings of the form `[base][+-]duration` or `base` will be expanded to the appropriate date (e.g. `tomorrow-2d` will get expanded to the date of yesterday).

`Base` can be one of today, tomorrow, monday, tuesday, wednesday, thursday, friday, saturday, sunday and `duration` is a valid duration as defined by [QQL](qql.md#Durations)

When the date type is used with the `humanize` option, dates are printed like "in 2 days", instead of the full `YYYY-MM-dd` format

## Duration Type

When the duration type is set it will be checked if the tag value is a valid duration according to [QQL](qql.md#Durations).
