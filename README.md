# Scrawl

Scrawl is a highly minimal ms paint clone. There are no colours, shapes, text
or anything like that. It's just a window that can be drawn on. Once closed,
the image is saved.

## Installation

### Prerequisites

1. `CGO_ENABLED=1` (should be on by default, but some distros disable it)
2. A C compiler (gcc/clang)
    - Anything that's been released in the last few years should be sufficient
3. libsdl, specifically [SDL2](https://wiki.libsdl.org/SDL2/FrontPage).
    - This is easiest to install via your package manager (apt, pacman, dnf,
      (*etc.*))

### From GitHub

```bash
go install github.com/sof202/scrawl@latest
```

### From source code

```bash
git clone https://github.com/sof202/scrawl
go install scrawl
```

## Usage

Run the program with:

```
scrawl path/to/image.png
```

The output file will be a png and can be given as either a relative path or
absolute path.

Once the program loads:

- Click+drag -> Draw lines
- Scroll -> Change brush size
- C -> Clear drawing
- ESCAPE -> Exit without saving

Closing the program will save the image in its current state to the path
provided on the command line.

## Motivation

I'm a mathematician, so as you'd expect, I like making notes with pencil and
paper still. However, more and more of my notes get taken digitally. When I
started taking notes digitally I'd do one of the following:

- Use Windows' notepad
- Take a picture of physical notes and upload it

These both had their obvious problems. Notepad's are obvious, the latter sounds
like it should work, but you can't search over your notes easily.

When I started working, I needed to move away from these approaches. Here's two
shitty ideas that others convinced me to use:

- Microsoft OneNote
    - The allure being that all notes are organised into directory like
      structures and it is 'easy' to add images, lists, tables (*etc.*)
- Obsidian (and all the other shit that's the same)
    - It uses markdown and has some of the same allure of OneNote.

Obsidian was my note taking app of choice for quite some time. It made a lot of
sense to use as I was using markdown in all of my code bases already, so it
felt very natural. The major problem I had with it was that I couldn't sync
across devices without paying (greedy bastards).

One day, I had an epiphany. Why was I using a notes app at all? I use a way
better tool every day. An actual fucking text editor (I use neovim by the way).
I felt stupid how obvious a solution this was. I didn't need some additional
app to take notes, I could just continue writing markdown in my preferred text
editor. This came with the bonus that I could now version control my notes and
sync them across devices (incredible). The biggest issue was rendering markdown
that had images or mermaid diagrams (*etc.*). This was easily fixed with the
use of plugins that are made by people that are much smarter than me. With
that, note taking apps were dead to me.

But there's still one last thing that annoyed me. When taking notes on paper,
I can make quick drawings to display things that words can't. Of course, I
could open up some drawing (web)app, take a screenshot (or export) and then
add this into my markdown. However... that sucks. That sounds like way too much
effort and also the images would likely be way too large for a simple doodle.

The final piece to the puzzle is scrawl. Alongside some code in my neovim
configuration, I can use a shortcut to open up a new window, make a quick
drawing and then save and add that into the document I'm writing. The addition
into the document is also automatable. In my opinion, this is as good as I can
get it without using some specialised app that allows you to draw where you
type (which almost certainly won't have vim bindings, so I shan't be having
it).

So that's it. To emulate pencil and pen:

- Version controlled notes written in markdown via a text editor
- Images and diagrams can be rendered by plugins
- Doodles can be handled by scrawl

### Extra

I've also been learning Go as of late, so this was a good opportunity to apply
what I've learned (and use SDL).
