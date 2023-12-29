import type { Component } from 'solid-js';
import { For, createResource, createSignal, Switch, Match } from 'solid-js';
import { A } from '@solidjs/router';
import './Catalog.css';
import type { Collectable } from '../resources';

import { TextField, Chip } from '@suid/material';

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

export const Catalog: Component = () => {
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
            <CatalogList collectables={collectables()!} />
          </Match>
        </Switch>
      </div>
    </div >
  );
};


type CatalogListProps = {
  collectables: Collectable[],
}

const CatalogList: Component<CatalogListProps> = ({ collectables }) => {
  const [filter, setFilter] = createSignal("")

  const filtered = () => {
    return collectables.filter((item) => {
      if (filter() === "") {
        return true;
      }
      let f = filter().toLowerCase();
      return item.name.toLowerCase().includes(f) || item.author.name.toLowerCase().includes(f);
    });
  }

  return (
    <>
      <div>
        <TextField label="Search" variant="outlined" autoComplete="off" fullWidth onChange={(e) => { setFilter(e.target.value); }} />
      </div>
      <div>
        <table>
          <thead>
            <tr>
              <th>Image</th>
              <th>Name</th>
              <th>Author</th>
              <th>Rarity</th>
            </tr>
          </thead>
          <tbody>
            <For each={filtered()} >
              {(item) => (
                <tr class={item.rarity}>
                  <td><A href={`/catalog/${item.id}`}><img src={item.image_url} style="width:80px;" /></A></td>
                  <td><A href={`/catalog/${item.id}`}>{item.name}</A></td>
                  <td><A href={`/user/${item.author.id}`}>{item.author.name}</A></td>
                  <td><Chip label={item.rarity} color="primary" /></td>
                </tr>
              )}
            </For>
          </tbody>
        </table>
      </div>
    </>
  );
}
