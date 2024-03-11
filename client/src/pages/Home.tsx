import type { Component, ResourceFetcher } from 'solid-js';
import { createResource, Switch, Match, For, onMount, onCleanup } from 'solid-js';
import { A, useNavigate } from '@solidjs/router';
import { Motion } from '@motionone/solid';
import { LoginButton } from '../components/LoginButton';
import type { IssuedCollectable } from '../resources';
import { rarityclass } from '../resources';
import { UserSearch } from '../components/UserSearch';
import { TimeAgo } from '../components/TimeAgo';
import { Button } from "@suid/material";

import './Home.css';

const fetchLatest: ResourceFetcher<true, IssuedCollectable[], unknown> = async (_, { value, refetching }) => {
  let req: Promise<Response>
  if (refetching && value !== undefined && value.length > 0) {
    let d = new Date(value[0].issued_at);

    req = fetch("/api/latest?since=" + d.getTime());
  } else {
    req = fetch("/api/latest");
  }

  let response = await req;
  if (response.status !== 200) {
    throw new Error("unexecpted status code: " + response.statusText);
  }

  let collectables = await response.json().then((resp) => {
    return resp;
  });

  if (value !== undefined) {
    return [...collectables, ...value]
  }
  return collectables;
}

export const Home: Component = () => {
  const [latestPulls, { refetch }] = createResource(fetchLatest)
  const navigate = useNavigate();

  let pollHandle: number | undefined = undefined;
  onMount(() => {
    pollHandle = setInterval(() => {
      refetch();
    }, 30 * 1000);
  });

  onCleanup(() => {
    if (pollHandle !== undefined) {
      clearInterval(pollHandle);
    }
  });

  return (
    <div class="split">
      <section class="intro">
        <h1>Shindaggers</h1>
        <p>Gaze upon your collection of hard earned <a href="https://twitch.tv/shindigs">Shindigs</a> Brand Knives.</p>

        <Button href="/catalog" variant="contained" color="primary" size="large">
          View The Collection
        </Button>

        <div>
          <h3>Lookup a collection:</h3>
          <UserSearch placeholder="username" onUserSelected={(u) => { u !== null && navigate(`/user/${u.id}`) }} />
        </div>

        <div><h3>Or</h3></div>

        <LoginButton />
      </section>
      <section class="pulls">
        <h2>Latest Pulls</h2>

        <Switch>
          <Match when={latestPulls()}>
            <For each={latestPulls()}>
              {(item) => (
                <A href={`/knife/${item.instance_id}`}>
                  <Motion.div
                    class={`pull ${rarityclass(item.rarity)}`}
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    transition={{ duration: 1, easing: 'ease-in-out' }}
                  >
                    <div class="image">
                      <img src={`https://images.shindaggers.io/images/${item.image_path}`} />
                    </div>
                    <div class="name">{item.name}</div>
                    <div class="info">
                      <div>PULLED BY {item.owner.name}</div>
                      <div class="time"><TimeAgo timestamp={item.issued_at} /></div>
                    </div>
                  </Motion.div>
                </A>
              )}
            </For>
          </Match>
          <Match when={latestPulls.loading} > <div>Loading...</div></Match>
          <Match when={latestPulls.error}><div>{latestPulls.error.toString()}</div></Match>

        </Switch>
      </section >
    </div >
  );
}

