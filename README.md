> Semi-automatic version control for Ableton Live™ projects

Hey, so I'm actually working on music again and this thing was broken af.

Often when I'm working on music I'll change something and later be
dissatisfied with it but not remember what I had or worse, not remember
what I changed.

At the same time, you want to be saving your work constantly in case of
a crash or power outage. Ableton does have a Crash Recovery Feature, 
which is nice but it also loses your Undo history every time you save,
which isn't so nice.

Version control is the commonly accepted solution and this thing
will `git commit` every time you hit `ctrl+s`.

## Recommended Usage

1. Run it from the folder where your project lives

    > Open it from the command line.

    > Should work on macOS now.

    > Windows users are recommended to use Git Bash for Windows.

2. Save often

    > Every time you save, the changes you made will be automatically committed to a git repository.

3. Tag versions occasionally by typing directly into the console

    > with something like "bassline" or "changed drums" just to give you an idea of where you were at

## New Stuff

Additionally, there are some new commands, just type them directly into the console:

 - **log**: print a list of changes

 - **checkout [hash]**: revert current project file.
 
    > To reload the project file in ableton, from the menu:

    > File > Open Recent File > [current project]

    > 
    
    > **when you check out a previous version, auto-committing is disabled until you use "save" or "cancel" commands, even if you restart the application!**

 - **cancel**: go back to the latest version of the project file

 - **save**: make the currently checked out version *the latest* version of the project file

    > effectively undo, but with the benefit of still being able to go back to the version(s) you don't want right now at a later time.

 - **set**: target the given filename as the current project

    > this is just for me to test this thing out please don't use this command

## WIP

This is still very much a work in progress, proper README coming soon™, possibly a GUI who knows

If you know how to use git at all, check out [my other thing](https://github.com/generaltso/go-git-em-tiger) which works very well in tandem with this.

But if my software isn't good enough for you, here's all you need to know about ableton project files:

```
#
# convert als into text version
#
cat project.als | gunzip > project.xml

#
# convert back to als
#
cat project.xml | gzip > project.als
```

That's it. 

btw this is still wtfpl so whatever kid just get out of my face B^U

.
