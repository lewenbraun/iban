# Iban indicator

This Quickshell configuration displays two click-through indicators on every
connected Wayland monitor. It uses the `overlay` layer, does not reserve screen
space, and does not receive keyboard focus.

## Installation

Install Quickshell on Arch Linux:

```sh
sudo pacman -S quickshell
```

Copy this directory into the user Quickshell configuration directory and start
it once from Hyprland:

```sh
mkdir -p ~/.config/quickshell
cp -r overlay ~/.config/quickshell/iban
```

```ini
exec-once = qs -n -d -c iban
```

`qs -n` prevents duplicate indicator processes. To apply QML changes, restart
the configuration:

```sh
qs kill -c iban
qs -n -d -c iban
```

## States

Both points use the same color and behavior:

- `recording`: muted red while audio is being recorded.
- `transcribing`: yellow while ElevenLabs is transcribing and the text is
  being copied or typed.
- `done`: green after successful delivery. It remains visible for
  `states.done.duration_ms` and disappears immediately when another recording
  starts.

Failures return to `idle`, so no completion color is shown.

## Configuration

Edit `~/.config/quickshell/iban/indicator.json`.

`indicator.shape` accepts `circle` or `square`. Each item in
`indicator.points` accepts one of `top-left`, `top-right`, `bottom-left`, or
`bottom-right` in `corner`. `x_px` is the horizontal inset from that corner;
`y_px` is the vertical inset. Both are signed logical pixels and can be changed
independently.

Every entry in `states` accepts `color`, `blink`, and `blink_interval_ms`.
`states.done.duration_ms` controls how long the green state remains visible.
