import type { Component } from 'solid-js';
import { createSignal, createResource, Switch, Match, For } from 'solid-js';
import { A, useNavigate } from '@solidjs/router';
import { Motion } from '@motionone/solid';
import { LoginButton } from './LoginButton';
import type { IssuedCollectable } from './resources';
import { rarityclass } from './resources';
import { UserSearch } from './Admin';

import './Home.css';

const fetchLatest = async (): Promise<IssuedCollectable[]> => {
  let response = await fetch("/api/latest");
  if (response.status !== 200) {
    throw new Error("unexecpted status code: " + response.statusText);
  }
  return await response.json().then((resp) => {
    if (resp.Collectables === undefined) { throw new Error("unexpected data format"); }
    return resp.Collectables;
  });
}

const rtfl = new Intl.RelativeTimeFormat('en', { numeric: 'always', style: 'long' });

const timeAgo = (dstr: string): string => {
  let d = new Date(dstr);
  let now = Date.now();

  let delta = now - d.getTime();

  if (delta < (1000 * 60)) {
    return "Just now";
  }

  if (delta < (1000 * 60 * 60)) {
    let minutes = Math.round(delta / (1000 * 60));
    return rtfl.format(-minutes, 'minute');
  }

  if (delta < (1000 * 60 * 60 * 24)) {
    let hours = Math.round(delta / (1000 * 60 * 60));
    return rtfl.format(-hours, 'hour');
  }

  let days = Math.round(delta / (1000 * 60 * 60 * 24));
  return rtfl.format(-days, 'day');
}

export const Home: Component = (props) => {
  const [latestPulls] = createResource(fetchLatest)
  const navigate = useNavigate();

  const [search, setSearch] = createSignal("");

  const handleKeyPress = (event: KeyboardEvent) => {
    if (event.key === 'Enter') {
      if (search() !== "") {
        navigate("/user/" + search());
      }
    }
  };

  const lookup = (event: MouseEvent) => {
    if (search() !== "") {
      navigate("/user/" + search());
    }
  };

  return (
    <div class="split">
      <section class="intro">
        <h1>Shindaggers</h1>
        <p>Gaze upon your collection of hard earned <a href="https://twitch.tv/shindigs">Shindigs</a> Brand Knives.</p>

        <A href="/catalog">
          <div class="button">
            View the Collection
          </div>
        </A>

        <div>
          <h3>Lookup a collection:</h3>
          <div class="input-button">
            <UserSearch placeholder="username" onUserSelected={(u) => { u !== null && navigate(`/user/${u.id}`) }} />
          </div>
        </div>

        <div><h3>Or</h3></div>

        <LoginButton />
      </section>
      <section class="pulls">
        <h2>Latest Pulls</h2>

        <Switch>
          <Match when={latestPulls.loading}><div>Loading...</div></Match>
          <Match when={latestPulls.error}><div>{latestPulls.error.toString()}</div></Match>
          <Match when={latestPulls()}>
            <For each={latestPulls()}>
              {(item) => (
                <A href={`/knife/${item.instance_id}`}>
                  <Motion.div
                    class={`pull ${rarityclass(item.rarity)}`}
                    animate={{ opacity: [0, 1] }}
                    transition={{ duration: 1, easing: 'ease-in-out' }}
                  >
                    <div class="image">
                      <img src={`https://images.shindaggers.io/images/${item.image_path}`} />
                    </div>
                    <div class="name">{item.name}</div>
                    <div class="info">
                      <div>PULLED BY {item.owner.name}</div>
                      <div class="time" title={item.issued_at}>{timeAgo(item.issued_at)}</div>
                    </div>
                  </Motion.div>
                </A>
              )}
            </For>
          </Match>

        </Switch>
      </section >
    </div>
  );
}

