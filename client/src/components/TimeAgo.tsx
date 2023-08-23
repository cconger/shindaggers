import type { Component } from 'solid-js';
import { createSignal } from 'solid-js';

export type TimeAgoProps = {
  timestamp: string | Date;
}


const rtfl = new Intl.RelativeTimeFormat('en', { numeric: 'always', style: 'long' });
const [ticker, setTicker] = createSignal(Date.now())

setInterval(() => {
  // Every minute, update the ticker
  setTicker(Date.now())
}, 60000)


export const TimeAgo: Component<TimeAgoProps> = (props) => {
  const timeAgo = () => {
    let d: Date;
    if (typeof props.timestamp === "string") {
      try {
        d = new Date(props.timestamp)
      } catch {
        return "Invalid Date";
      }
    } else {
      d = props.timestamp;
    }
    let now = ticker();

    let delta = now - d.getTime();
    if (isNaN(delta)) {
      return "Invalid Date";
    }

    if (delta < (1000 * 60)) {
      return "Just now";
    }

    if (delta < (1000 * 60 * 60)) {
      let minutes = Math.round(delta / (1000 * 60));
      return rtfl.format(-minutes, 'minute');
    }

    if (delta < (1000 * 60 * 60 * 24)) {
      let hours = Math.round(delta / (1000 * 60 * 60));
      return rtfl.format(-hours, 'hour');
    }

    let days = Math.round(delta / (1000 * 60 * 60 * 24));
    return rtfl.format(-days, 'day');
  }

  const timestamp = () => {
    if (typeof props.timestamp == "string") {
      return props.timestamp;
    }
    return props.timestamp.toISOString();
  }

  return (
    <span title={timestamp()}>{timeAgo()}</span>
  )
}
