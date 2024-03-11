import type { Component } from 'solid-js';
import { Show, createResource, Switch, Match } from 'solid-js';
import { A, useParams } from '@solidjs/router';
import { MiniCard } from '../components/MiniCard';
import './Catalog.css';
import type { IssuedCollectable, User, UserDuelStats } from '../resources';
import { DistributionChart } from '../components/Chart';
import { ListingFromCollectables, UserCollectionList } from '../components/UserCollectionList';

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
            <h1>{usercollection()!.User.name}</h1>
          </div>
          <Show when={usercollection()!.Equipped}>
            <div class="catalog-equipped">
              <h2>Equipped</h2>
              <A href={`/knife/${usercollection()!.Equipped!.instance_id}`}>
                <MiniCard collectable={usercollection()!.Equipped!} />
              </A>
            </div>
          </Show>
        </div>
        <div class="catalog">
          <UserCollectionList collection={usercollection()?.Collectables || []} />
        </div>
      </Match>
    </Switch>
  );
};


type PercentageProps = {
  numerator: number;
  denominator: number;
}

export const Percentage: Component<PercentageProps> = (props) => {
  let value = () => {
    return (props.numerator / props.denominator) * 100;
  }

  return (
    <Show when={props.denominator !== 0} fallback="-">
      {value().toFixed(1)}%
    </Show>
  )
}
