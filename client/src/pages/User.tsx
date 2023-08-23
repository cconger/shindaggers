import type { Component } from 'solid-js';
import { Show, For, createResource, Switch, Match } from 'solid-js';
import { A, useParams } from '@solidjs/router';
import { MiniCard, MiniListing } from '../components/MiniCard';
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
    <div class="user">
      <Switch>
        <Match when={usercollection.loading}>
          <div>Loading...</div>
        </Match>
        <Match when={usercollection.error}>
          <div>Error loading catalog</div>
        </Match>
        <Match when={usercollection()}>
          <UserProfile user={usercollection()!} />
        </Match>
      </Switch>
    </div>
  );
};


const UserProfile = (props: { user: UserCollection }) => {
  return (
    <div>
      <div class="header">
        <h1>{props.user.User.name}</h1>
      </div>
      <div class="profile">
        <section class="profile-card">
          <Show when={props.user.Equipped}>
            <div class="title">Equipped</div>
            <div class="content">
              <MiniListing collectable={props.user.Equipped!} />
            </div>
          </Show>
        </section>
        <section class="profile-card double">
          <div class="title">Duels</div>
          <div class="content">
            <div class="duel-stats">
              <div class="match-history">
                <div class="match won"></div>
                <div class="match-trace"></div>
                <div class="match lost"></div>
                <div class="match-trace"></div>
                <div class="match tied"></div>
                <div class="match-trace"></div>
                <div class="match won"></div>
                <div class="match-trace"></div>
                <div class="match won"></div>
                <div class="match-trace"></div>
                <div class="match won"></div>
                <div class="match-trace"></div>
                <div class="match lost"></div>
              </div>
              <div class="stats">
                <div class="stat">
                  <div class="stat-name">Wins</div>
                  <div class="stat-value">4</div>
                </div>
                <div class="stat">
                  <div class="stat-name">Ties</div>
                  <div class="stat-value">1</div>
                </div>
                <div class="stat">
                  <div class="stat-name">Losses</div>
                  <div class="stat-value">2</div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section class="profile-card double">
          <div class="title">Collection</div>
          <div class="content">
            <div class="collection-list">
              <div class="controls">
                <label>
                  <input type="checkbox" checked />
                  Show Duplicates
                </label>
              </div>

              <For each={props.user.Collectables}>
                {(collectable) => (
                  <MiniListing collectable={collectable} />
                )}
              </For>
            </div>
          </div>
        </section>

        <section class="profile-card">
          <div class="title">Stats</div>
          <div class="content">
            Chart goes here...
          </div>
        </section>
      </div>
    </div>
  )
}


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
