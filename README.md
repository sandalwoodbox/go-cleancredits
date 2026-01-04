# go-cleancredits

A tool for removing on-screen text from video

## Installation

Navigate to the latest release at https://github.com/sandalwoodbox/go-cleancredits/releases
and download the package for your platform. The package names represent the platform used for
compiling the binary; there may be cross-compatibility. For example, I can run
macos-15-intel-x86_64 on MacOS 12.

### From source

Requirements:

- [git](https://git-scm.com/)
- [golang](https://go.dev/doc/install)
- OpenCV ([MacOS](https://gocv.io/getting-started/macos/), [Linux](https://gocv.io/getting-started/linux/), [Windows](https://gocv.io/getting-started/windows/))

```bash
git clone https://github.com/sandalwoodbox/go-cleancredits.git
cd go-cleancredits
go build
./go-cleancredits
```

## Usage

Click "Open video file" and select the video file you want to open.

There are three tabs: Mask, Draw, and Render. There is also a preview area to
the right of the tabs. These are all described in the following sections.

### Mask

This tab allows you to create a mask - pixels that will be removed and "inpainted" (that
is, replaced with colors chosen by nearby pixels). The mask should aim to
include all of the text you want to remove and nothing else.

The Mask tab has the following controls:

1. **Layer.** Allows adding up to 5 mask layers that will be stacked on top
   of each other to determine the final mask. This is useful, for example, if
   there are multiple colors of text that you want to remove at once.
   * TODO: Add layer support.
2. **Mode.** Select whether the current layer should determine places that
   will always / never be inpainted. Each layer will override the layers underneath it.
   * TODO: Add mode widget.
3. **Frame.** The frame to use when building the mask for the current layer.
   This will be displayed in the preview area.
4. **Hue / Saturation / Value.** Set what ranges of colors should be
   considered for the current mask layer.
   * TODO: Add color picker widget 
5. **Grow.** Add additional pixels to the edge of the current mask layer's
    selected areas. This can be useful to ensure that video compression
    artifacts don't negatively impact the inpainting process.
6. **Crop.** Select what areas of the frame will be considered for the current
   mask layer. This can be useful if other parts of the image have similar
   colors to the text you want to remove. Crop is applied after HSV + Grow.

![Screenshot of Mask tab GUI](/screenshots/mask.png)

### Draw

TODO: Add Draw support

This tab allows you to "draw" manual overrides to force specific areas to always
/ never be inpainted. This will be applied after all mask layers.

The Draw tab has the following controls:

1. **Frame.** Select the frame to display in the preview area.
2. **Mode.** Whether the drawn pixels should be always inpainted, never inpainted,
   or if the overrides should be reset.
3. **Size.** "Paintbrush" size.

![Screenshot of Draw tab GUI](/screenshots/draw.png)

### Render

This tab allows you to render a specific portion of the video with your options applied.

The Render tab has the following controls:

1. **Start frame.** The first frame to be rendered. This will be displayed in
   the Preview area when modified.
2. **End frame.** The last frame to be rendered (inclusive).  This will be
   displayed in the Preview area when modified.
3. **Inpaint radius.** How many neighboring pixels to use to calculate the
   right color for each pixel. The larger this number is, the slower rendering
   will be. Use the "Preview" view mode to see what the result will
   look like.
4. **Render.** Choose an output target and render the inpainted result.

![Screenshot of Render tab GUI](/screenshots/render.png)

### Preview

The Preview area renders the currently visible frame.

The Preview area has the following controls:

1. **View.** How to render the visible frame. The values have the following meanings:
   * Areas to inpaint. Display the areas that will be inpainted - that is, the final mask - assuming that the current layer is the final layer. For example, if you have layer 3 selected, this will take layers 1 & 2 into account but not layers 4 & 5.
   * Overrides. Show only the overrides layer (modified in the Draw tab.)
   * Preview. Display what this frame would look like if inpainted. This mode will be slower to render.
   * Original. Show the original frame.
2. **Zoom.** Modify the zoom level.
3. **Anchor X / Y.** Modify the point that the zoom will center on.

## Profiling

1. **CPU profiling:** `go run . -cpuprofile=cpu.prof`
2. **Memory profiling:** `go run . -memprofile=mem.prof`
   * This will dump a memory profile after you exit the program.
3. **OpenCV Mat profiling:** `go build -tags matprofile && ./go-cleancredits`
   * Use this if memory is growing quickly and nothing is visible in the memory
     profile. OpenCV Mats are created by C and therefore aren't visible in Go
     memory profiling.
4. **Live profiling:** `go run . -profserver`
   * Access the server at http://localhost:6060/debug/pprof/

### RoyaltyFreeVideos license

Footage used in screenshots and tests is courtesy of [RoyaltyFreeVideos](https://www.youtube.com/c/RoyaltyFreeVideos/about). At the time the footage was pulled, this page read:

```
Please feel free to use any of the videos on this channel for personal and commercial use without credit and without payment in your video projects. The only thing you may not do with our videos is re-upload them in raw format outside of your own creations. A credit or link back to our channel is appreciated though is not necessary.
```

Videos used:

- [Horses In Field](https://www.youtube.com/watch?v=ieI8DDNLBgs)