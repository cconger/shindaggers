import type { Component } from 'solid-js';
import { For, createResource, createSignal, Switch, Match } from 'solid-js';
import { A } from '@solidjs/router';
import './Catalog.css';
import type { Collectable } from '../resources';
import { Rarity, rarities } from '../resources';
import { MiniCard } from '../components/MiniCard';

import { TextField, Chip, Stack } from '@suid/material';

import styles from '../components/UserCollectionList.module.css';

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

type RarityFilters = {
  [key in Rarity]?: boolean;
}

const CatalogList: Component<CatalogListProps> = ({ collectables }) => {
  const [filters, setFilters] = createSignal<RarityFilters>({
    [Rarity.Common]: true,
    [Rarity.Uncommon]: true,
    [Rarity.Rare]: true,
    [Rarity.SuperRare]: true,
    [Rarity.UltraRare]: true,
  });
  const [filter, setFilter] = createSignal("")

  const textFiltered = () => {
    return collectables.filter((item) => {
      if (filter() === "") {
        return true;
      }
      let f = filter().toLowerCase();
      return item.name.toLowerCase().includes(f) || item.author.name.toLowerCase().includes(f);
    });
  }

  const filtered = () => {
    return textFiltered().filter((item) => {
      return filters()[item.rarity];
    });
  }

  const rarityCounts = () => {
    let counts = {
      [Rarity.Common]: 0,
      [Rarity.Uncommon]: 0,
      [Rarity.Rare]: 0,
      [Rarity.SuperRare]: 0,
      [Rarity.UltraRare]: 0,
    };
    textFiltered().forEach((item) => {
      counts[item.rarity as Rarity]++;
    });
    return counts;
  }

  return (
    <div class={styles.CatalogList}>
      <div class={styles.Filters}>
        <Stack direction="column" spacing={1}>
          <TextField label="Search" variant="outlined" autoComplete="off" fullWidth onChange={(e) => { setFilter(e.target.value); }} />
          <Stack direction="row" spacing={1}>
            <For each={rarities}>
              {(rarity) => (
                <Chip
                  onClick={() => { setFilters((f) => ({ ...f, [rarity]: !filters()[rarity] })) }}
                  color={filters()[rarity] ? "primary" : "default"}
                  label={rarity + " (" + rarityCounts()[rarity] + ")"}
                />
              )}
            </For>
          </Stack>
        </Stack>
      </div>
      <div class={styles.Catalog}>
        <For each={filtered()} >
          {(item) => (
            <A href={`/catalog/${item.id}`}>
              <MiniCard collectable={item} />
            </A>
          )}
        </For>
      </div>
    </div>
  );
}
