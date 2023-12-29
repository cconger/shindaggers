import type { Component, JSX } from 'solid-js';

import styles from './Tooltip.module.css';

type TooltipProps = {
  tip: string;
  children: JSX.Element;
}

export const Tooltip: Component<TooltipProps> = (props) => {

  return (
    <div class={styles.tooltip}>
      {props.children}
      <div class={styles.tip}>
        <p> {props.tip} </p>
      </div>
    </div>
  );
}
