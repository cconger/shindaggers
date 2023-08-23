import type { Component } from "solid-js";
import { A } from '@solidjs/router';
import { createResource, Match, Switch, For } from "solid-js";
import { useParams } from "@solidjs/router";
import type { User, IssuedCollectable, UserDuelStats } from "../resources";
import { TimeAgo } from "../components/TimeAgo";

import './Event.css';

type EventPayload = {
  leaderboard: {
    user: User;
    stats: UserDuelStats;
  }[];
  collectables: { [id: string]: IssuedCollectable };
  last_fights: Fight[];
}

type Fight = {
  id: string;
  user_ids: string[];
  collectable_ids: string[];
  outcomes: number[];
  time: string;
}

const getEventDetails = async (slug: string): Promise<EventPayload> => {
  let resp = await fetch("/api/event/" + slug + "/stats");
  if (resp.status !== 200) {
    throw new Error("Error fetching event stats");
  }
  return await resp.json();
}

export const Event: Component = () => {
  const params = useParams();

  const [event] = createResource(() => params.slug, getEventDetails);

  return (
    <Switch>
      <Match when={event.loading}>loading...</Match>
      <Match when={event.error}>error</Match>
      <Match when={event()}>
        <EventDash event={event()!} />
      </Match>
    </Switch>
  )
}

// outcomeclass generates a classname based on integer outcome
const outcomeclass = (outcome: Number): string => {
  switch (outcome) {
    case 1:
      return "win";
    case 0:
      return "tie";
    case -1:
      return "lose";
    default:
      return "";
  }
}

export const EventDash: Component<{ event: EventPayload }> = (props) => {

  let userMap = () => {
    const m = new Map<string, User>();
    for (const entry of props.event.leaderboard) {
      m.set(entry.user.id, entry.user);
    }
    return m;
  };

  return (
    <div class="event">
      <div>
        <h1>Leaderboard</h1>
        <div class="leaderboard">
          <div class="leaderboard-entry">
            <div class="leaderboard-header">
              Name
            </div>
            <div class="leaderboard-header">
              Wins - Ties - Losses
            </div>
          </div>
          <For each={props.event.leaderboard}>
            {(entry) => (
              <div class="leaderboard-entry">
                <div class="name"><A href={`/user/${entry.user.id}`}>{entry.user.name}</A></div>
                <div class="stats">
                  <span class="stat wins">{entry.stats.wins}</span>
                  -
                  <span class="stat ties">{entry.stats.ties}</span>
                  -
                  <span class="stat losses">{entry.stats.losses}</span>
                </div>
              </div>
            )}
          </For>
        </div>
      </div>
      <div>
        <h1>Matches</h1>
        <For each={props.event.last_fights}>
          {(fight) => (
            <div>
              <div> <TimeAgo timestamp={fight.time} /> </div>
              <div class="fight-entry">
                <div class={`participant left ${outcomeclass(fight.outcomes[0])}`}>
                  <div class="name">{userMap().get(fight.user_ids[0])?.name || "Unknown"}</div>
                  <img src={props.event.collectables[fight.collectable_ids[0]]?.image_url || "https://images.shindaggers.io/images/default.png"} />
                </div>

                <div class={`participant right ${outcomeclass(fight.outcomes[1])}`}>
                  <div class="name">{userMap().get(fight.user_ids[1])?.name || "Unknown"}</div>
                  <img src={props.event.collectables[fight.collectable_ids[1]]?.image_url || "https://images.shindaggers.io/images/default.png"} />
                </div>
              </div>
            </div>
          )}
        </For>
      </div>
    </div>
  );
}
