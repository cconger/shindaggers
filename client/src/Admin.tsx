import type { Component, JSX } from 'solid-js';
import { Show, For, Match, Switch, createResource, createSignal, on } from 'solid-js';
import { Rarity, rarities } from './resources';
import type { Collectable, User } from './resources';
import { createStore } from 'solid-js/store';
import { Card } from './Card';
import { useAuthManager } from './LoginButton';
import { Button } from './Button';

import './Admin.css';

type UploadState = {
  name: string,
  rarity: Rarity,
  author: null | User,
  image: null | File,
  imagePreview: string,
  preview: boolean,
}

export const AdminPage: Component = () => {



  return (
    <div>
      <h1>Admin Page</h1>

      <h2>Functionalty</h2>
      <ul>
        <li>Create new knife</li>
        <li>Modify knives</li>
        <li>Disable knives</li>
        <li>Delete knife</li>
        <li>Issue Knife</li>
        <li>Change Weights</li>
      </ul>

      <CollectableForm />

    </div>
  )
}

const CollectableForm: Component = () => {
  let fileInputRef: HTMLInputElement | undefined = undefined;

  const [store, setStore] = createStore<UploadState>({
    name: "Placeholder",
    rarity: Rarity.Common,
    author: null,
    image: null,
    imagePreview: "",
    preview: false,
  })

  let collectable = () => {
    let name = store.name;
    let author = store.author || {
      id: "",
      name: "undefined",
    };

    return {
      id: "undefined",
      name: name,
      author: author,
      rarity: store.rarity,
      image_path: "",
      image_url: store.imagePreview,
    };
  }

  let handleImageChange: JSX.EventHandler<HTMLInputElement, Event> = (e) => {
    const files = e.currentTarget.files;
    if (files === null) { return; }
    const image = files[0];
    const imagePreview = URL.createObjectURL(image);
    setStore({
      image,
      imagePreview,
      preview: true,
    });
  }

  let am = useAuthManager();

  let handleUpload = async () => {
    if (!am.token()) {
      console.error("not logged in, cannot upload image")
      return
    }

    if (store.image === null) {
      console.error("image is null, cannot upload")
      return
    }

    const formdata = new FormData();
    formdata.append("image", store.image);

    let resp = await fetch("/api/image", {
      body: formdata,
      method: 'POST',
      headers: {
        "Authorization": am.token()!,
      },
    })

    if (!resp.ok) {
      throw new Error("unexpected status code:" + resp.status);
    }

    return await resp.json();
  }

  let handleCreateCardAPI = async () => {
    return null;
  }

  let handleExistingImage: JSX.EventHandler<HTMLInputElement, Event> = (e) => {
    let url = e.currentTarget.value;

    if (store.image) {
      URL.revokeObjectURL(store.imagePreview);
      setStore({ image: null })
    }

    setStore({ imagePreview: "https://images.shindaggers.io/images/" + url, preview: true })
  }

  let handleCreateCard = async () => {
    let res = await handleUpload();
  }
  return (
    <div class="form">
      <h2>New Knife</h2>
      <div class="input-button">
        <input placeholder="Name" onInput={(e) => setStore({ name: e.currentTarget.value })} />
      </div>
      <div>
        <select class="selector" id="rarity" name="rarity" onChange={(e) => setStore({ rarity: e.currentTarget.value as Rarity })}>
          <For each={rarities}>
            {(r) => (
              <option value={r}>{r.toString()}</option>
            )}
          </For>
        </select>
      </div>
      <UserSearch onUserSelected={(u) => setStore({ author: u })} />
      <div class="image-selector">
        <div class="input-button">
          <input placeholder="Existing Image" onChange={handleExistingImage} />
        </div>
        <div><h3>Or</h3></div>
        <div>
          <input type="file" hidden ref={fileInputRef} accept="image/*" onChange={handleImageChange} />
          <div class="button" onClick={() => fileInputRef?.click()}>Upload Image</div>
        </div>
      </div>

      <Show when={store.preview} keyed>
        <Button text="Create Card" onClick={handleCreateCard} />
        <h3>Preview</h3>
        <Card collectable={collectable()} />
      </Show>
    </div>
  )
}

type UserSearchProps = {
  placeholder?: string,
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
  const [valid, setValid] = createSignal(false);
  const [searchResults] = createResource(() => search(), searchUsers)

  let placeholder = props.placeholder || "User"

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
    "input-button": true,
    "valid": valid(),
    "invalid": !valid(),
  });

  return (
    <>
      <div classList={cls()}>
        <input ref={inputEl} placeholder={placeholder} onInput={(e) => { setValid(false); setSearch(e.target.value); }} />
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

export const AdminCatalog: Component = () => {
  return (
    <div>AdminCatalog</div>
  )
}

export const AdminKnife: Component = () => {
  return (
    <div>AdminKnife</div>
  )
}
