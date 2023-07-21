import type { Component } from 'solid-js';
import { createResource, Switch, Match } from 'solid-js';
import { useParams, A } from '@solidjs/router';
import { Card } from './Card';
import type { IssuedCollectable } from './resources';

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

export const Pull: Component<PullProps> = (props) => {
  const params = useParams();

  const [collectable] = createResource(() => params.id, fetchIssuedCollectable);

  const issuedAt = (): string => {
    let c = collectable();
    if (c === undefined) { return "" };
    var d = new Date(c.issued_at);
    return d.toLocaleString();
  }

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
