import type { Component } from 'solid-js';
import type { AdminCollectable } from './resources';
import { createSignal, Show } from 'solid-js';
import { RequireLogin, CollectableForm } from './Admin';
import { Rarity, User } from './resources';
import { useAuthManager } from './LoginButton';

export const Creator: Component = () => {
  let am = useAuthManager();

  return (
    <section class="admin-page">
      <h2>You think you're a bad enough dude to make your own knife?</h2>
      <h3>Prove it</h3>
      <RequireLogin fallback={<h1>You gotta be logged in to create</h1>}>
        <CreatorForm user={am.user()!} />
      </RequireLogin>
    </section>
  )
}


type CreatorFormProps = {
  user: User;
}

const createCollectable = async (c: AdminCollectable): Promise<AdminCollectable> => {
  let am = useAuthManager();

  let resp = await fetch("/api/collectable", {
    method: "POST",
    body: JSON.stringify({ Collectable: c }),
    headers: {
      "Content-Type": "application/json",
      "Authorization": am.token()!,
    },
  });

  if (!resp.ok) {
    throw new Error("unexpected status code: " + resp.status);
  }

  let body = await resp.json();
  return body.Collectable;
}

const CreatorForm: Component<CreatorFormProps> = (props) => {
  const [msg, setMsg] = createSignal<string | null>(null);

  const defaultKnife: AdminCollectable = {
    id: "",
    name: "Knife",
    rarity: Rarity.Common,
    image_path: "default.png",
    image_url: "https://images.shindaggers.io/images/default.png",
    author: props.user,
    deleted: false,
    approved: false,
  };

  let submit = async (c: AdminCollectable) => {
    try {
      let res = await createCollectable(c)
      setMsg(`Your knife "${res.name}" is pending approval!`)
    } catch (e) {
      setMsg(`${e}`)
    }
  }

  return (
    <>
      <Show when={msg()}>
        <div class="flex-mid"><h3>{msg()}</h3></div>
      </Show>
      <CollectableForm collectable={defaultKnife} onSubmit={submit} authuser preview />
    </>
  )
}

