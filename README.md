# drift

_being a piece of textmode visual art for contemplating the dead months_

everything will be lost<br>
like fragile flakes<br>
on a winter wind<br>

![the above lines drifting down a terminal as snowflakes](./drift.gif)

this is a program that accepts lines on `STDIN`. then, it turns each character of each line into a snowflake.

they pile up and then melt away. gusts of wind will occasionally push them.

you can hit enter to trigger gusts.

## installation

```
go install github.com/vilmibm/drift
```

note this installs to `~/go/bin/drift`. you may want to put `~/go/bin` on your `$PATH`.

## usage

pipe something in.

```
cat poem.txt | drift
```

quit whenever by pressing `Esc` or hitting `ctrl+c`

## author

nate smith <vilmibm@protonmail.com>
