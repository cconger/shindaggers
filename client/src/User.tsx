import type { Component } from 'solid-js';
import { Show, For, createResource, Switch, Match } from 'solid-js';
import { A, useParams } from '@solidjs/router';
import { MiniCard } from './MiniCard';
import './Catalog.css';
import type { IssuedCollectable, User } from './resources';
import { DistributionChart } from './Chart';

import './User.css';

type UserCollection = {
  User: User;
  Collectables: IssuedCollectable[];
  Equipped: IssuedCollectable | null;
}

const fetchUserCollection = async (id: string): Promise<UserCollection> => {
  let response = await fetch(`/api/user/${id}/collection`)
  if (response.status !== 200) {
    throw new Error("unexpected status code " + response.statusText);
  }
  return await response.json();
}

export const UserCollection: Component = (props) => {
  const params = useParams();
  const [usercollection] = createResource(() => params.id, fetchUserCollection)

  let total = () => {
    let collection = usercollection();
    if (collection === undefined) { return 0; }
    return collection.Collectables.length;
  }

  return (
    <Switch>
      <Match when={usercollection.loading}>
        <div>Loading...</div>
      </Match>
      <Match when={usercollection.error}>
        <div>Error loading catalog</div>
      </Match>
      <Match when={usercollection()}>
        <div class="catalog-header">
          <div class="catalog-title">
            <h1>{usercollection()!.User.name}'s Collection</h1>
            <h2>{total()} Knives</h2>
          </div>
          <Show when={usercollection()!.Equipped}>
            <div class="catalog-equipped">
              <h2>Equipped</h2>
              <MiniCard collectable={usercollection()!.Equipped!} />
            </div>
          </Show>
          <div class="catalog-stats">
            <DistributionChart collection={usercollection()!.Collectables} />
          </div>
        </div>
        <div class="catalog">
          <For each={usercollection()!.Collectables} >
            {(item) => (
              <A href={`/knife/${item.instance_id}`}>
                <MiniCard collectable={item} />
              </A>
            )}
          </For>
        </div>
      </Match>
    </Switch>
  );
};
