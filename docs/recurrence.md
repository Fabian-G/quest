# Recurrence

The minimal configuration to enable recurrence is:

```toml
[recurrence]
rec-tag = "rec"
```

This configuration will use the `rec` tag to determine the recurrence
interval and the `t` and `due` tag to determine the threshold and due-tag.

This implementation of recurring tasks aims to be compatible with other todo.txt clients like pter and simpletask. 
So for further details checkout the [pter docs](https://vonshednob.cc/pter/documentation.html#recurring-tasks)