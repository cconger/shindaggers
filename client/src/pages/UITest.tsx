import type { Component } from 'solid-js';
import { Button } from '../components/Button';
import { TimeAgo } from '../components/TimeAgo';

export const UITest: Component = (props) => {
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
    <div>
      <h1>UI Test</h1>

      <section>
        <h2>TimeAgo</h2>
        <div style="display: flex; flex-direction: column;">
          <div>
            <TimeAgo timestamp={(new Date())} />
          </div>
        </div>
      </section>

      <section>
        <h2>Buttons</h2>
        <div style="display: flex; flex-direction: column;">
          <Button text="Click to Succeed" onClick={success} />
          <Button text="Click to Fail" onClick={fail} />
          <Button text="Click to Spin" onClick={spin} />
        </div>
      </section>
    </div>
  );
}
