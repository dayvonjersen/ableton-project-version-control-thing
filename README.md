<!--
    should mention that you 
        LOSE YOUR ENTIRE FUCKING UNDO HISTORY EVERY TIME YOU SAVE IN ABLETON

    that it should work on macOS now
        but the GUI will be for windows only at this time
         (maybe should use Qt instead of walk to scaffold...
            nah, i'm familiar with walk so it'll be easier to mock up
            then we can make the real thing with Qt
         )

    describe workflow for using previous versions:
    FIRST:
        - git log, copy hash of revision you want
            highlight how using commit messages really helps in this regard
            if you save often (which you should)
    
    PREVIEW (optional):
        NOTE: auto-commit will be disabled!!!

        - git checkout [hash]
            - can also do git checkout HEAD^, HEAD^^^^^ etc...
        - File > Open Recent Files > current project

    REVERT:
        git checkout master
        git ls-tree [hash]

        100644 blob 1237b90f9e287a8ba1ea55ea489c655b204bd7f0    test.xml
        100644 blob f1e6b766d28bf993837a6ce375b634b5ab58cbfd    this is a test for arbitrary filenames %% &.xml

        NOTE: .ALS NOT .XML
        git cat-file [file hash] | gzip > filename.als

        (should auto-commit)    
-->
# ableton-autosave

> Automatic version control with git for Ableton Live projects

It's common knowledge that Ableton Live project files are just gzipped XML.

Thus, one can simply:

```bash
cat my-awesome-song.als | gunzip > my-awesome-song.xml
```

And then put the XML file under version control.

This program automates this process. (and alleviates the need for `my-awesome-song Copy.als`, `my-awesome-song Copy Copy.als`, `my-awesome-song Copy Copy.als` ...)

**(windows only atm)**

*This is a glorified shell script I made for my own personal use and has nothing to do with Ableton the company and is not an official product use at your own peril not responsible for any loss of data or audio fidelity Ableton Live is probably a registered trademark all rights reserved pls do not kil*

Users with knowledge of git and Go are encouraged to look at (and possibly contribute to) the source code!

## Pre-requisites

 - [Git for Windows](https://git-for-windows.github.io/)

 - [Go](https://golang.org)

## Installation

```
go install github.com/generaltso/ableton-autosave
```

## Usage

```
$ cd /Directory_Where_You_Keep_Your_Projects
$ ableton-autosave.exe
2017/08/05 18:07:56 watching .
Initialized empty Git repository in [ ... ]
[master (root-commit) ... ]
[master ... ]
 1 file changed, 4139 insertions(+)
 create mode 100644 my-awesome-song.xml
2017/08/05 18:07:56 watching my-awesome-song Project

[ ... ]

Type commit messages in the running console after you save and hit enter
[master ... ] Type commit messages in the running console after you save and hit enter
 Date: Sat Aug 5 18:20:21 2017 -0400
 1 file changed, 4139 insertions(+)
```

1. Run `ableton-autosave` (from the command line) in the directory where your Ableton project(s) are stored.

    - this will initialize one empty repository per project folder and commit the existing file(s).

    - additionally, it will add a git hook for `git checkout` which will automatically update the `.als` file so you can open a previous version straight away.

2. Every time you `Save` (`ctrl+s`) in Ableton while `ableton-autosave` is running, it will commit your changes (with an empty message)

    - if you type in the console window while `ableton-autosave` is running and hit the return key, you can change the commit message to something useful

    - if you create a new project in the directory it will `git init` automatically.

3. You can now use standard git commands to navigate your project's evolution!

    - you should not have the file open in Ableton while reverting to previous versions...

    - `ableton-autosave` will *not* auto-commit when not on the `master` branch

    - `ableton-autosave` will *only* manage `.als` files and not any preset files, audio files or any other part of your project.

    - please note the `git checkout` hook described above



## License

wtfpl
