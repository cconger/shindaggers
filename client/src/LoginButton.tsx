import type { Component, Setter, Accessor, Resource, JSX } from 'solid-js';
import { A, Navigate } from '@solidjs/router';
import { Show, Switch, Match } from 'solid-js';
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
  if (response.status === 403 || response.status === 404) {
    // Your token is no good...
    useAuthManager().setToken(null);
    return null
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
    [this.user] = createResource(() => this.token() || "", fetchUser);

    createEffect(() => {
      let t = this.token();
      if (t === null) {
        localStorage.removeItem(lsKey);
      } else {
        localStorage.setItem(lsKey, t);
      }
    });
  }
}

let am: AuthManager | undefined;
export const useAuthManager = () => {
  if (am === undefined) {
    am = new AuthManager();
  }
  return am;
}

export const NavLogin: Component = () => {
  let am = useAuthManager();

  return (
    <Switch>
      <Match when={am.user.loading}>
        <a href="#">-</a>
      </Match>
      <Match when={am.user() === null}>
        <a href="/oauth/login">Login</a>
      </Match>
      <Match when={am.user()}>
        <A href={`/user/${am.user()!.id}`}>{am.user()!.name}</A>
        <a onClick={() => am.setToken(null)}>Logout</a>
      </Match>
    </Switch>
  )
}

export const IfLoggedIn: Component<{ children: JSX.Element }> = (props) => {
  let am = useAuthManager();

  return (
    <Show when={am.token()}>
      {props.children}
    </Show>
  );
}

export const LoginButton: Component = () => {
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

export const LoginLander: Component = () => {
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
