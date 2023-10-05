# Styling

You can define styles on two levels: line level or tag level.
Line level styles are defined in the main section of your config file like this:

```toml
styles = [
	{ if = 'priority >= prioC', fg = "3" },
]
```

This would change the foreground color of the whole line to color 3 if the corresponding task has a priority larger than C.

Tag styles on the other hand are defined within a tag definition:

```toml
[tags.due]
type = "date"
humanize = true
styles = [
	{ if = 'date(tag(it, "due"), maxDate) <= today + 1d', fg = "1" },
]
```

This will color only the due tag column in color 1 if the task is due today or tomorrow.

If multiple style definitions match an item, the first match will be applied. 
If a line style matches it will always override any tag style.

Colors can be either defined by using the ANSI color or with the 3 byte hex RGB color (e.g. #00FF00).