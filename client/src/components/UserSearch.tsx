import type { Component } from 'solid-js';
import { For, Match, Switch, createResource, createSignal } from 'solid-js';
import type { User } from '../resources';
import { TextField } from '@suid/material';

import './UserSearch.css';

type UserSearchProps = {
  placeholder?: string,
  default?: User,
  onUserSelected(u: User | null): unknown,
}

const searchUsers = async (search: string): Promise<User[]> => {
  if (search == "") {
    return [];
  }

  let resp = await fetch("/api/users?search=" + search)
  if (!resp.ok) {
    throw new Error("unexpected status")
  }

  let users = await resp.json();

  return users.Users;
}

export const UserSearch: Component<UserSearchProps> = (props) => {
  const [search, setSearch] = createSignal("");
  const [valid, setValid] = createSignal(!!props.default);
  const [searchResults] = createResource(() => search(), searchUsers)

  let inputEl: HTMLInputElement | undefined = undefined;
  let selectUser = (user: User) => {
    props.onUserSelected(user);
    setValid(true);
    setSearch("");
    if (inputEl !== undefined) {
      inputEl.value = user.name;
    }
  };

  let cls = () => ({
    "user-complete": true,
    "valid": valid(),
    "invalid": !valid(),
  });

  return (
    <>
      <div classList={cls()}>
        <TextField ref={inputEl} label="Search" autoComplete="off" variant="outlined" fullWidth onChange={(e) => { setValid(false); setSearch(e.target.value); }} />
        <div class="results">
          <Switch>
            <Match when={searchResults.loading}><img src="https://images.shindaggers.io/images/spinner.svg" /></Match>
            <Match when={searchResults.error}><div>{searchResults.error}</div></Match>
            <Match when={searchResults()}>
              <For each={searchResults()}>
                {(user) => (
                  <div onClick={() => selectUser(user)}>{user.name}</div>
                )}
              </For>
            </Match>
          </Switch>
        </div>
      </div>
    </>
  );
}
