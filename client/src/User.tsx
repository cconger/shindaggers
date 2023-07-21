import type { Component } from 'solid-js';
import { For, createResource, Switch, Match } from 'solid-js';
import { A, useParams } from '@solidjs/router';
import MiniCard from './MiniCard';
import './Catalog.css';
import type { IssuedCollectable, User } from './resources';

type UserCollection = {
  User: User;
  Collectables: IssuedCollectable[];
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

  return (
    <Switch>
      <Match when={usercollection.loading}>
        <div>Loading...</div>
      </Match>
      <Match when={usercollection.error}>
        <div>Error loading catalog</div>
      </Match>
      <Match when={usercollection()}>
        <h1>{usercollection()!.User.name}'s Collection</h1>
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
