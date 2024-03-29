import type { Component } from 'solid-js';
import { createSignal, Switch, Match } from 'solid-js';
import { Motion, Presence } from "@motionone/solid";

import "./Button.css";

export type ButtonProps = {
  text: string;
  danger?: boolean
  warn?: boolean
  onClick?: () => Promise<unknown>;
};

enum ButtonState {
  Default = 0,
  Pending,
  Success,
  Fail,
}

export const Button: Component<ButtonProps> = (props) => {
  const [state, setState] = createSignal(ButtonState.Default);

  const handleClick = async () => {
    if (props.danger) {
      let res = window.confirm("Are you sure?")
      if (!res) { return };
    }
    if (state() !== ButtonState.Default) {
      return false;
    }
    setState(ButtonState.Pending);

    if (props.onClick) {
      try {
        await props.onClick();
        setState(ButtonState.Success)
      } catch {
        setState(ButtonState.Fail)
      }
    }
  };

  let initial = {
    y: 58,
    opacity: 0,
  };

  let animate = {
    y: 0,
    opacity: 1,
  };

  let exit = {
    opacity: 0,
    y: 58,
  };

  let transition = {
    duration: 0.15
  };

  let cls = {
    "button": true,
    "button-action": true,
    "warn": props.warn,
    "danger": props.danger,
  };

  return (
    <div classList={cls} onClick={handleClick}>
      <Presence>
        <Switch>
          <Match when={state() === ButtonState.Default}>
            <Motion.div
              animate={animate}
              exit={exit}
              transition={transition}
            >
              {props.text}
            </Motion.div>
          </Match>
          <Match when={state() === ButtonState.Pending}>
            <Motion.div
              initial={initial}
              animate={animate}
              exit={exit}
              transition={transition}
            >
              <img src="https://images.shindaggers.io/images/spinner.svg" alt="spinner" />
            </Motion.div>
          </Match>
          <Match when={state() === ButtonState.Success}>
            <Motion.div
              initial={initial}
              animate={animate}
              exit={exit}
              transition={transition}
            >
              Success
            </Motion.div>
          </Match>
          <Match when={state() === ButtonState.Fail}>
            <Motion.div
              initial={initial}
              animate={animate}
              exit={exit}
              transition={transition}
            >
              Fail
            </Motion.div>
          </Match>
        </Switch>
      </Presence>
    </div>
  );
}
