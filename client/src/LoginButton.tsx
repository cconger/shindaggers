import type { Component, Setter, Accessor, Resource } from 'solid-js';
import { A, Navigate } from '@solidjs/router';
import { Switch, Match } from 'solid-js';
import { createSignal, createResource, createEffect } from 'solid-js';

import type { User } from './resources';

const lsKey = "authtoken";

type authinfo = {
  id: string;
  name: string;
  token: string;
}

const fetchUser = async (token: string | null): Promise<User | null> => {
  if (token === null || token === "") {
    return null;
  }
  let response = await fetch("/api/user/me", {
    headers: {
      'Authorization': token,
    },
  });
  if (response.status === 403) {
    // Invalid token
    useAuthManager().logout();
    return null;
  }
  if (response.status !== 200) {
    throw new Error("unexpected status code " + response.statusText);
  }
  return await response.json().then((res) => res.User);
}

class AuthManager {
  user: Resource<User | null>
  authInfo: authinfo | undefined;
  token: Accessor<string | null>;
  setToken: Setter<string | null>;

  constructor() {
    const rawToken = localStorage.getItem(lsKey);
    [this.token, this.setToken] = createSignal(rawToken);
    [this.user] = createResource(() => this.token(), fetchUser);

    createEffect(() => {
      let t = this.token();
      if (t === null) {
        localStorage.removeItem(lsKey);
      } else {
        localStorage.setItem(lsKey, t);
      }
    });
  }

  logout() {
    this.setToken(null);
  }
}

let am: AuthManager | undefined;
const useAuthManager = () => {
  if (am === undefined) {
    am = new AuthManager();
  }
  return am;
}

export const NavLogin: Component = (props) => {
  let am = useAuthManager();

  return (
    <Switch>
      <Match when={am.user.loading}>
        <a href="#">-</a>
      </Match>
      <Match when={am.user.error}>
        <a href="/oauth/login">Login</a>
      </Match>
      <Match when={am.user() === null}>
        <a href="/oauth/login">Login</a>
      </Match>
      <Match when={am.user()}>
        <A href={`/user/${am.user()!.id}`}>{am.user()!.name}</A>
      </Match>
    </Switch>
  )
}

export const LoginButton: Component = (props) => {
  let am = useAuthManager();

  return (
    <Switch>
      <Match when={am.user.loading}>
        <a href="/oauth/login">
          <div class="button">
            Login with Twitch
          </div>
        </a>
      </Match>
      <Match when={am.user() === null}>
        <a href="/oauth/login">
          <div class="button">
            Login with Twitch
          </div>
        </a>
      </Match>
      <Match when={am.user()}>
        <A href={`/user/${am.user()!.id}`}>
          <div class="button">
            {am.user()!.name}'s Collection
          </div>
        </A>
      </Match>
    </Switch>
  )
}

export const LoginLander: Component = (props) => {
  let am = useAuthManager();

  let url = new URL(window.location.href);
  let params = new URLSearchParams(url.hash.slice(1));
  let token = params.get("token");
  if (token) {
    am.setToken(token);
  }

  return (
    <Switch>
      <Match when={am.user.loading}>
        <div>Loading User</div>
      </Match>
      <Match when={am.user.error}>
        <div>{am.user.error.toString()}</div>
      </Match>
      <Match when={am.user()}>
        <Navigate href={`/user/${am.user()!.id}`} />
      </Match>
    </Switch>
  );
}
