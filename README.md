# gip
List git commits across grouped repositories within a given time period

# Vague command plans

``` sh
gip
  view (group) [time]
  viewrepo (repo) [time]
  repos
    add (gitdir)
    list [group]
    remove (repo) [group]
  groups
    create (name)
    list
    add (repo) (group)
    rename (group) (name)
    remove (group)
```

# Bad example

``` sh
$ gip repos list
# No repos yet!
  
$ gip groups list
# No groups yet!
  
$ gip repos add ~/my/wrok/repo
# No .git directory found at /Users/xyz/my/wrok/repo

$ gip repos add ~/my/work/repo
# Repository at /Users/xyz/my/work/repo added as "work-site" to the "all" group

$ gip repos add ~/my/other/repo personal
# Repository at /Users/xyz/my/other/repo added as "p-site" to the "personal" group

$ gip repos list
# p-site (/Users/xyz/my/other/repo)
# work-site (/Users/xyz/my/work/repo)

$ gip groups list
# all (2 repositories)
# personal (1 repository)

$ gip repos list personal
# p-site (/Users/xyz/my/other/repo)

$ gip groups rename personal private
# Group "personal" renamed to "private"

$ gip groups remove private
# Do you also want to remove all groups within this group from the "all" group? (Y/n) Y

$ gip groups create golibs
# Created group "golibs"

$ gip groups list
# all (1 repository)
# golibs (0 repositories)

$ gip repos remove work-site
# Doing this will remove "work-site" from all groups. Are you sure? (Y/n) n

$ gip groups add work-site work
# Repository at /Users/xyz/my/work/repo added as "work-site" to the "work" group

$ gip repos remove work-site work
# Repository at /Users/xyz/my/work/repo ("work-site") removed from the "work" group

$ gip view all
# ...logs for "all" group from 12am...

$ gip view all 1.day
# ...logs for "all" group from the last 24 hours...

$ gip viewrepo work-site
# ...logs from the "work-site" repo from 12am...
```
