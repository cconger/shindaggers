import type { Component } from 'solid-js';
import { createResource, Show, Switch, Match } from 'solid-js';
import { useParams, A } from '@solidjs/router';
import { Card } from './Card';
import type { IssuedCollectable } from './resources';
import { useAuthManager } from './LoginButton';
import { Button } from './Button';

import './Pull.css';

type PullProps = {
}

const fetchIssuedCollectable = async (id: string): Promise<IssuedCollectable> => {
  let response = await fetch("/api/issued/" + id)
  if (response.status === 404) {
    throw new Error("IssuedCollectable does not exist")
  }
  return await response.json().then((resp) => {
    if (resp.Collectable) return resp.Collectable;
    throw new Error("Unexpected Datatype")
  })
}

const equipKnife = async (token: string, userID: string, knifeID: string) => {
  let resp = await fetch("/api/user/equip", {
    method: "POST",
    headers: {
      "Authorization": token,
    },
    body: JSON.stringify({
      UserID: userID,
      IssuedID: knifeID,
    }),
  });

  if (resp.status !== 200) {
    throw new Error("Error")
  }
};

export const Pull: Component<PullProps> = (props) => {
  const params = useParams();

  const [collectable] = createResource(() => params.id, fetchIssuedCollectable);

  const issuedAt = (): string => {
    let c = collectable();
    if (c === undefined) { return "" };
    var d = new Date(c.issued_at);
    return d.toLocaleString();
  }

  const am = useAuthManager();

  const equip = async () => {
    let owner = collectable()?.owner.id || "";
    let instance_id = collectable()?.instance_id || "";
    let token = am.token() || "";
    return await equipKnife(token, owner, instance_id);
  };

  return (
    <Switch>
      <Match when={collectable.loading}>
        <div>Loading</div>
      </Match>
      <Match when={collectable.error}>
        <div>{collectable.error.toString()}</div>
      </Match>
      <Match when={collectable()}>
        <Card collectable={collectable()!} issuedCollectable={collectable()!} />
        <section class="info-card">
          <Show when={am.user() && am.user()!.id == collectable()!.owner.id}>
            <div class="flex-mid">
              <Button text="Equip this Knife" onClick={equip} />
            </div>
          </Show>
          <div>
            <div class="info-card-header">Owner</div>
            <div class="info-card-body">
              <A href={`/user/${collectable()!.owner.id}`}>{collectable()!.owner.name}</A>
            </div>
          </div>
          <div>
            <div class="info-card-header">Acquired At</div>
            <div class="info-card-body">
              {issuedAt()}
            </div>
          </div>
        </section>
      </Match>
    </Switch>
  );
}
