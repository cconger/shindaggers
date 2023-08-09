import type { Component, JSX } from 'solid-js';
import { Show, For, Match, Switch, createResource, createSignal } from 'solid-js';
import { A, useParams } from '@solidjs/router';
import { Rarity, rarities } from './resources';
import type { AdminCollectable, Collectable, User } from './resources';
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
    <div class="admin-page">
      <h1>Admin Page</h1>

      <h2>New Knife</h2>
      <CollectableForm preview />

      <h2>Collectables</h2>
      <CollectableList />
    </div>
  )
}


const fetchAdminCollectables = async (): Promise<AdminCollectable[]> => {
  let am = useAuthManager();

  let resp = await fetch("/api/admin/collectables", {
    headers: {
      "Authorization": am.token()!,
    },
  });

  if (!resp.ok) {
    throw new Error("unexpected status code: " + resp.status)
  }

  let body = await resp.json()

  return body.Collectables || [];
}

const CollectableList: Component = () => {
  const [collectables] = createResource(fetchAdminCollectables)

  const [filterDeleted, setFilteredDeleted] = createSignal(true);

  const shown = () => {
    let cs = collectables();
    if (!cs) {
      return [];
    }

    let res = cs.slice().reverse();

    if (filterDeleted()) {
      res = res.filter((c) => !c.deleted);
    }

    return res;
  }

  return (
    <div>
      <div>
        <label>Show Deleted
          <input
            type="checkbox"
            checked={!filterDeleted()}
            onChange={(e) => setFilteredDeleted(!e.currentTarget.checked)}
          />
        </label>
      </div>
      <Switch>
        <Match when={collectables.loading}>
          <div>Loading...</div>
        </Match>
        <Match when={collectables.error}>
          <div>Error: {collectables.error.toString()}</div>
        </Match>
        <Match when={collectables()}>
          <div class="collectable-table">
            <div class="header">ID</div>
            <div class="header">Name</div>
            <div class="header">Author</div>
            <div class="header">Rarity</div>
            <div class="header">Image</div>
            <div class="header">Active</div>
            <For each={shown()}>
              {(collectable) => (
                <>
                  <div> <A href={`/admin/knife/${collectable.id}`}>{collectable.id}</A> </div>
                  <div> <A href={`/admin/knife/${collectable.id}`}>{collectable.name}</A> </div>
                  <div>{collectable.author.name}</div>
                  <div>{collectable.rarity}</div>
                  <div><A href={collectable.image_url}>{collectable.image_path}</A></div>
                  <div>{collectable.deleted ? "❌" : "✅"}</div>
                </>
              )}
            </For>
          </div>
        </Match>
      </Switch>
    </div>
  )
}


const deleteKnife = async (id: string): Promise<AdminCollectable> => {
  let am = useAuthManager();

  let resp = await fetch("/api/admin/collectable/" + id, {
    method: "DELETE",
    headers: {
      "Authorization": am.token()!,
    },
  });

  if (!resp.ok) {
    throw new Error("unexpected status code: " + resp.status);
  }

  let body = await resp.json();
  return body.Collectable;
}

type CollectableFormProps = {
  collectable?: AdminCollectable;
  preview?: boolean;
  allowDelete?: boolean;
  onSubmit?: (c: Collectable) => Promise<unknown>
}

const CollectableForm: Component<CollectableFormProps> = (props) => {
  let fileInputRef: HTMLInputElement | undefined = undefined;

  const [store, setStore] = createStore<UploadState>({
    name: props.collectable?.name || "",
    rarity: props.collectable?.rarity || Rarity.Common,
    author: props.collectable?.author || null,
    image: null,
    imagePreview: props.collectable?.image_url || "",
    preview: !!props.collectable,
  })

  let collectable = () => {
    let name = store.name;
    let author = store.author || {
      id: "",
      name: "undefined",
    };

    return {
      id: props.collectable?.id || "",
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

  let handleExistingImage: JSX.EventHandler<HTMLInputElement, Event> = (e) => {
    let url = e.currentTarget.value;

    if (store.image) {
      URL.revokeObjectURL(store.imagePreview);
      setStore({ image: null })
    }

    setStore({ imagePreview: "https://images.shindaggers.io/images/" + url, preview: true })
  }

  let handleSubmit = async () => {
    await handleUpload();
    if (props.onSubmit) {
      await props.onSubmit(collectable());
    } else {
      console.error("no handler for submit", collectable())
    }
  }

  return (
    <div class="form">
      <div class="input-button">
        <input
          placeholder="Name"
          value={store.name}
          onInput={(e) => setStore({ name: e.currentTarget.value })}
        />
      </div>
      <div>
        <select class="selector" id="rarity" name="rarity" onChange={(e) => setStore({ rarity: e.currentTarget.value as Rarity })}>
          <For each={rarities}>
            {(r) => (
              <option value={r} selected={store.rarity == r}>{r.toString()}</option>
            )}
          </For>
        </select>
      </div>
      <UserSearch
        default={props.collectable?.author}
        onUserSelected={(u) => setStore({ author: u })}
      />
      <div class="image-selector">
        <div class="input-button">
          <input
            placeholder="Existing Image"
            value={props.collectable?.image_path || ""}
            onChange={handleExistingImage}
          />
        </div>
        <div class="or"><h3>Or</h3></div>
        <div>
          <input type="file" hidden ref={fileInputRef} accept="image/*" onChange={handleImageChange} />
          <div class="button" onClick={() => fileInputRef?.click()}>Upload Image</div>
        </div>
      </div>

      <Show when={store.preview} keyed>
        <div class="controls">
          <Button text="Submit" onClick={handleSubmit} />
          <Show when={props.allowDelete && props.collectable}>
            <Show when={props.collectable?.deleted}><h3>KNIFE IS DELETED</h3></Show>
            <Show when={!props.collectable?.deleted}>
              <Button text="Delete" danger onClick={() => deleteKnife(props.collectable!.id)} />
            </Show>
          </Show>
        </div>
        <h3>Preview</h3>
        <Show when={props.preview}>
          <Card collectable={collectable()} />
        </Show>
      </Show>
    </div>
  )
}

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
        <input ref={inputEl} value={props.default?.name || ""} placeholder={placeholder} onInput={(e) => { setValid(false); setSearch(e.target.value); }} />
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

const fetchAdminCollectable = async (id: string): Promise<AdminCollectable> => {
  let am = useAuthManager();

  let resp = await fetch("/api/admin/collectable/" + id, {
    headers: {
      "Authorization": am.token()!,
    },
  });

  if (!resp.ok) {
    throw new Error("unexpected status code: " + resp.status);
  }

  let body = await resp.json();
  return body.Collectable;
}

export const AdminKnife: Component = () => {
  const params = useParams();

  let [collectable] = createResource(() => params.id, fetchAdminCollectable)

  return (
    <div class="admin-page">
      <h1>AdminKnife</h1>
      <Switch>
        <Match when={collectable.loading}>
          <div>Loading</div>
        </Match>
        <Match when={collectable.error}>
          <div>Error</div>
        </Match>
        <Match when={collectable()}>
          <h3>Modify Knife:</h3>
          <CollectableForm collectable={collectable()!} allowDelete preview />
        </Match>
      </Switch>
    </div>
  )
}
