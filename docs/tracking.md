# Time Tracking

Time tracking in quest is delegated to your local installation of timewarrior.
To enable time tracking you have to:

1. Tell quest which tag to use to be able to remember what currently is being tracked.
2. Have timewarrior installed.

The following config snippet will set the tracking tag to `tr`:

```toml
[tracking]
tag = "tr"
```

## Starting a task

First make sure that timewarrior is installed. You can verify that quest can find it by running: `quest version`.

You can now set the `tr` tag on a task to any value you like. For example:

```bash
quest set tr:active on 42
```
The actual value does not matter, it is only important that quest notices a change in the value of the tag.
After that you will notice that timewarrior will have started tracking task 42
by using the projects, contexts and the description (without any projects/contexts/tags) as tags:

```bash
~ ❯ timew
Tracking "Add tracking chapter" +quest
  Started 2024-03-07T10:55:04
  Current            11:01:17
  Total               0:06:13
```

For convenience there is also the `track` view command. Which effectively also
just sets the tracking tag, but additionally it makes sure that only one task
matches your selection (setting the tracking tag on multiple tasks will start
tracking **a** task, but it is undefined which one).
The `track` command is also much more convenient, because it doesn't make you think
about a value for the tracking tag, but it just sets it to a timestamp (minutes since epoch).
Therefore we could have started tracking like this as well:
```bash
quest track 42
```

## Changing an actively tracked task

Sometimes you might need to change the currently tracked task for whatever reason (spelling mistakes, missing project...).
In this case you can just edit the task as you normally would for example using set:
```bash
quest set @work on 42
```
After that you will notice that this change was automatically passed on to timewarrior:

```bash
~ ❯ timew
Tracking "Add tracking chapter" +quest @work
  Started 2024-03-07T10:55:04
  Current            11:12:18
  Total               0:17:14
```

Note that in order to not interfere with tasks you track with timewarrior without using quest
this only works if the current timewarrior tags exactly match the previous state of the task.
Therefore, if you manually run `timew start ...` or `timew tag ....` the currently active
tracking will be considered unrelated to a quest task.

## Stopping a task

To stop tracking a task you have several options.

The easiest method is to simply run `timew stop`. In this case the tracking tag will remain
on the previously tracked task. This is fine, though, because it will be cleared when you start
tracking another task. One thing to keep in mind is, that if you want to restart 
tracking on a task where the tracking tag is already set you'll have to set it to a different value.
If you just use `track` to start tracking this will be handled for you.

Another option is to clear the tracking tag: `quest unset tr on 42`.

And the last option is to complete the item, which also automatically stops tracking: `quest complete 42`.

For the purpose of stopping a task it is also very convenient to have a macro at hand that matches
the task that is actively being tracked. For example:

```toml
[[macro]]
name = "tracked"
args = ["item"]
result = "bool"
query = '!done(arg0) && !(tag(arg0, "tr") == "")'
inject-it = true
```

Now `quest complete tracked` will also do the trick.

## Configuring timewarrior tags

By default quest uses projects, contexts and the description without any projects/context/tags as timewarrior tags.
As of now this is not configurable.
What you can configure is whether or not project/context prefixes should be passed to timewarrior.

```toml
[tracking]
tag = "tr"
trim-project-prefix = true 
trim-context-prefix = true
```

In that case the previous example would have looked like this:

```bash
~ ❯ timew
Tracking "Add tracking chapter" quest work
  Started 2024-03-07T10:55:04
  Current            11:12:18
  Total               0:17:14
```

You can also configure which todo.txt tag should be included as timewarrior tags. 
This is useful when you want to include some metadata like a ticket id.

```toml
[tracking]
tag = "tr"
include-tags = ["ticket"]
```
