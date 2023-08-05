import type { Component } from 'solid-js';
import { createSignal, Switch, Match } from 'solid-js';
import { Motion, Presence } from "@motionone/solid";

import "./Button.css";
import spinner from './spinner.svg';

export type ButtonProps = {
  text: string;
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
    if (state() !== ButtonState.Default) {
      return false;
    }
    setState(ButtonState.Pending);

    if (props.onClick) {
      try {
        let res = await props.onClick();
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

  return (
    <div class="button button-action" onClick={handleClick}>
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
              <img src={spinner} alt="spinner" />
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

export const ButtonTest: Component = (props) => {
  const success = () => {
    return new Promise(resolve => setTimeout(resolve, 1000));
  }

  const fail = () => {
    return new Promise((_, rej) => setTimeout(rej, 1000));
  }

  const spin = () => {
    return new Promise(() => { });
  }

  return (
    <div style="display: flex; flex-direction: column;">
      <Button text="Click to Succeed" onClick={success} />
      <Button text="Click to Fail" onClick={fail} />
      <Button text="Click to Spin" onClick={spin} />
    </div>
  );
}
