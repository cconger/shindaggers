import { render } from 'solid-js/web';
import type { Component } from 'solid-js';
import { fetchIssuedCollectable } from './Pull';
import { MiniCard } from './MiniCard';
import { createResource, Show, Switch, Match } from 'solid-js';
import { rarityclass } from './resources';

import './overlay.css';

type OverlayProps = {
  id: string;
}

const Overlay: Component<OverlayProps> = (props) => {
  const [collectable] = createResource(() => props.id, fetchIssuedCollectable);

  return (
    <Switch>
      <Match when={collectable.loading}>
        <div></div>
      </Match>
      <Match when={collectable.error}>
        <div>Error</div>
      </Match>
      <Match when={collectable()}>
        <div class={`overlay ${rarityclass(collectable()!.rarity)}`}>
          <div><h1>{collectable()!.owner.name}'s</h1></div>
          <div><h2>earned a {collectable()!.rarity}</h2></div>
          <div class="flex-mid">
            <MiniCard collectable={collectable()!} />
          </div>
          <div class="info">
            <h2>
              Crafted by {collectable()!.author.name}
            </h2>
          </div>
        </div>
      </Match>
    </Switch>
  )
}

const root = document.getElementById('root');

const getID = (): string | null => {
  let url = new URL(window.location.href);
  const candidate = url.pathname.split('/').pop() || '';
  if (!isNaN(parseInt(candidate, 10))) {
    return candidate;
  }

  return url.searchParams.get('id');
}

const id = getID();

render(() => (
  <Show when={id !== null}>
    <Overlay id={id!} />
  </Show>
), root!);
