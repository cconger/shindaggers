import type { Component } from 'solid-js';
import { onMount } from 'solid-js';

const paintText = (c: HTMLCanvasElement, text: string, scale: number, lineHeight: number, fontSize: number) => {
  let strippedText = text.replace(/^\s*|\s(?=\s)|\s*$/g, "");
  let words = strippedText.split(' ');

  let ctx = c.getContext('2d');
  if (ctx == null) { return; }

  ctx.font = `800 ${fontSize}px Montserrat`;
  ctx.fillStyle = 'white';
  ctx.scale(scale, scale);

  let canvasWidth = c.width;
  let renderWidth = canvasWidth / scale;

  let word = words.shift();
  let lines: [string, number][] = [];
  while (word) {
    word = word.toUpperCase();
    let metrics = ctx.measureText(word);

    if (word.length > 3 && metrics.width > renderWidth) {
      // split and try again
      let mid = word.length / 2;
      words.unshift(word.slice(0, mid), word.slice(mid));
    } else {
      lines.push([word, metrics.width]);
    }

    word = words.shift();
  }

  // Start from top
  let cursorY = lineHeight;

  for (let [line, width] of lines) {
    let letters = line.split('');
    let cursorX = 0;
    let delta = (renderWidth - width) / (line.length - 1);
    for (let i = 0; i < letters.length; i++) {
      let metrics = ctx.measureText(letters[i]);
      ctx.fillText(letters[i], cursorX, cursorY);
      cursorX += metrics.width + delta;
    }
    cursorY += lineHeight;
  }
}

type TextArtBGProps = {
  name: string;
  lineHeight: number;
  size: {
    width: number;
    height: number;
  };
  fontSize?: number;
}

const TextArtBG: Component<TextArtBGProps> = (props) => {
  let dpr = window.devicePixelRatio || 1;

  let { width, height } = props.size

  let fontSize = props.fontSize || 30;


  let canvas: HTMLCanvasElement | undefined;
  onMount(() => {
    if (canvas !== undefined) {
      paintText(canvas, props.name, dpr, props.lineHeight, fontSize);
    }
  })

  return (
    <canvas
      width={width * dpr}
      height={height * dpr}
      ref={canvas}
      style={`width: ${width}px; height: ${height}px;`}
    >
    </canvas>
  );
};

export default TextArtBG;
