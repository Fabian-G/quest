# Configuration

This annotated example configuration shows all available options:

```toml
# The Path to the todo.txt quest should use by default. 
# Can be override with the -f option
# DEFAULT: $HOME/.local/share/quest/todo.txt
todo-file = "$HOME/path/to/your/todo.txt"

# The Path where quest should store completed items. 
# Done tasks are NOT moved automatically, but only when using the archive command.
# DEFAULT: $HOME/.local/share/quest/done.txt
done-file = "$HOME/path/to/your/done.txt"

# When set to a value > 0 quest will make a backup of the todo.txt/done.txt 
# whenever it is changed.
# Setting this to 10 will keep at most 10 backups.
# DEFAULT: 0
backup = 10

# If set to false Tags that are not declared in the tags section will emit an 
# error if encountered in the todo.txt
# DEFAULT: true
unknown-tags = false

# A list of tags that should be removed from a task upon completion
# DEFAULT: []
clear-on-done = ["do"]

# A list of style definitions that should be applied to a whole line in the task 
# view.
# DEFAULT: []
styles = [
	{ if = 'blocked', fg = "3" },
]

# The quest-score is a value calculated based on the urgency
# and priority of a task.
# High urgency and high priority results in a high quest-score.
# DEFAULT: empty (therefore disabled)
[quest-score]
# The date tags that define the urgency of a task.
# Might also be offset by a duration value.
# Only the first existing tag is considered for urgency calculation.
urgency-tags = [ "due", "t+3m" ]

# When a task will start to be considered as urgent (in days). 
urgency-begin = 45 

# If the urgency-tag is unset on a task the urgency will be calculated by
# adding this duration to the creation date.
# This is useful to avoid low priority tasks being left undone forever.
urgency-default = "3m"

# The minimal priority. All lower priorities are considered unimportant.
min-priority = "E" 

# Properties for configuring the timewarrior integration.
# The Projects, Contexts and description are used for tags
[tracking]
trim-project-prefix = false # If the + should be removed from projects
trim-context-prefix = false # If the @ should be removed from contexts

# Tag configuration for the recurrence feature
# DEFAULT: empty (therefore disabled)
[recurrence]
# Duration tag that defined the recurrence interval
rec-tag = "rec" 

# Date tag that defines the due date
due-tag = "due" 

# Date tag that defined the threshold date
threshold-tag = "t" 

# When this option is set to true an item that is spawned by completing a
# recurrent item will be assigned the same priority as the original.
preserve-priority = false

# List of tag definitions to enable tag expansions and styling
[tags]
[tags.due]
# type of the tag. One of "string", "date", "duration", "int"
type = "date" 

# If the tag is a date tag and humanize = true it will be output in a human 
# friendly format
humanize = true 

# Per tag style definitions. This will color only the corresponding cell.
styles = [
	{ if = 'date(tag(it, "due"), maxDate) <= today + 1d', fg = "1" },
	{ if = 'date(tag(it, "due"), maxDate) <= today + 3d', fg = "6" }
]

# Default view definition.
# This has 2 purposes. 
# 1. It is the view that is displayed when no view is specified.
# 2. The values set here serve as the defaults for other views
[default-view]
# A list of columns to be included in the output
projection = ["line", "done" ,"priority","projects","contexts","description"]  

# A list of tags, projects and contexts that should be removed 
# from the description column.
# This is useful to avoid displaying redundant information. 
# If we already have a whole column dedicated to projects
# we won"t need the project displayed inside the description column as well.
clean = ["@ALL","+ALL"]

# Whether or not this view should be opened in interactive mode with live reload
interactive = false

# When the user adds an item through this view the configured prefix gets 
# automatically prepended.
add-prefix = '@inbox'

# Limits the amount of tasks that will be outputted to 10.
# Set to -1 to show all tasks
limit = 10

# A view definition with the name inbox.
[views.inbox]
query = '!done && @inbox'
sort = ["+creation","+description"]
clean = ["@inbox"]
projection = ["line","creation","description"]

[[macro]]
# The name of the macro
name = "blocked" 

# The arguments of the macro
args = ["item"] 

# The expected result type of the macro
result = "bool" 

# The actual query that the macro performs
query = '!(tag(arg0, "after", "") == "") && (exists pre in items: tag(pre, "id", "") == tag(arg0, "after") && !done(pre))'

# Whether or not we want to enable the special it-injection syntax for this macro, 
# so that we can write blocked, instead of blocked(it)
inject-it = true
```
