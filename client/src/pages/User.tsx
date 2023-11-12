import type { Component } from 'solid-js';
import { Show, For, createResource, Switch, Match } from 'solid-js';
import { A, useParams } from '@solidjs/router';
import { MiniCard } from '../components/MiniCard';
import './Catalog.css';
import type { IssuedCollectable, User, UserDuelStats } from '../resources';
import { DistributionChart } from '../components/Chart';

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
            <h3>{total()} Knives</h3>
            <DuelStats user={usercollection()!.User} />
          </div>
          <Show when={usercollection()!.Equipped}>
            <div class="catalog-equipped">
              <h2>Equipped</h2>
              <A draggable = "false" href={`/knife/${usercollection()!.Equipped!.instance_id}`}>
                <MiniCard collectable={usercollection()!.Equipped!} />
              </A>
            </div>
          </Show>
          <div class="catalog-stats">
            <DistributionChart collection={usercollection()!.Collectables} />
          </div>
        </div>
        <div class="catalog">
          <For each={usercollection()!.Collectables} >
            {(item) => (
              <A draggable = "false" href={`/knife/${item.instance_id}`}>
                <MiniCard collectable={item} />
              </A>
            )}
          </For>
        </div>
      </Match>
    </Switch>
  );
};


const fetchDuelStats = async (id: string): Promise<UserDuelStats> => {
  let response = await fetch(`/api/user/${id}/stats`)
  if (response.status !== 200) {
    throw new Error("unexpected status code " + response.statusText);
  }
  let payload = await response.json();
  return payload.Stats;
}

type DuelStatsProps = {
  user: User;
}

export const DuelStats: Component<DuelStatsProps> = (props) => {

  const [userstats] = createResource(() => props.user.id, fetchDuelStats)

  let total = () => {
    let s = userstats()
    if (s === undefined) {
      return 0
    }
    return s.wins + s.losses + s.ties;
  }

  return (
    <Switch>
      <Match when={userstats.loading}>
        Loading Stats...
      </Match>
      <Match when={userstats.error}>
        <></>
      </Match>
      <Match when={userstats()}>
        <>
          <h2>Duel Stats </h2>
          <div class="stat-block">
            <div class="stat">
              <div class="header" title="Wins">W</div>
              <div class="count">{userstats()?.wins}</div>
              <div class="percent"><Percentage numerator={userstats()?.wins || 0} denominator={total()} /></div>
            </div>
            <div>-</div>
            <div class="stat">
              <div class="header" title="Losses">L</div>
              <div class="count">{userstats()?.losses}</div>
              <div class="percent"><Percentage numerator={userstats()?.losses || 0} denominator={total()} /></div>
            </div>
            <div>-</div>
            <div class="stat">
              <div class="header" title="Ties">T</div>
              <div class="count">{userstats()?.ties}</div>
              <div class="percent"><Percentage numerator={userstats()?.ties || 0} denominator={total()} /></div>
            </div>
          </div>
        </>
      </Match>
    </Switch>
  )
}

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
