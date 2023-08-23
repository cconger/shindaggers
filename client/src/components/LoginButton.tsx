import type { Component, JSX } from 'solid-js';
import { A, Navigate } from '@solidjs/router';
import { Show, Switch, Match } from 'solid-js';
import { useAuthManager } from '../auth';

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
