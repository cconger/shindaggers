import type { Component } from 'solid-js';
import { For, createResource, Switch, Match } from 'solid-js';
import { A } from '@solidjs/router';
import { MiniCard } from '../components/MiniCard';
import './Catalog.css';
import type { Collectable } from '../resources';

const fetchCatalog = async (): Promise<Collectable[]> => {
  let response = await fetch("/api/catalog")
  if (response.status !== 200) {
    throw new Error("unexpected status code " + response.statusText);
  }
  return await response.json().then((resp) => {
    if (resp.Collectables) return resp.Collectables;
    return [];
  });
}

export const Catalog: Component = (props) => {
  const [collectables] = createResource(fetchCatalog)

  return (
    <div>
      <h1>Catalog</h1>

      <div class="catalog">
        <Switch>
          <Match when={collectables.loading}>
            <div>Loading...</div>
          </Match>
          <Match when={collectables.error}>
            <div>Error loading catalog</div>
          </Match>
          <Match when={collectables()}>
            <For each={collectables()} >
              {(item) => ( 
                <A draggable="false" href={`/catalog/${item.id}`} >
                  <MiniCard collectable={item} />
                  </A>
              )}
            </For>
          </Match>
        </Switch>
      </div>
    </div>
  );
};