import type { Component } from 'solid-js';
import { Show, For, createSignal } from 'solid-js';
import type { IssuedCollectable, Collectable } from '../resources';
import { Rarity, rarities } from '../resources';
import { A } from '@solidjs/router';

import { TextField, Stack, Chip } from '@suid/material';
import Verified from '@suid/icons-material/Verified';
import Favorite from '@suid/icons-material/Favorite';
import Check from '@suid/icons-material/Check'

import styles from './UserCollectionList.module.css';

export const ListingFromCollectables = (issuedCollectables: IssuedCollectable[]) => {
  const byID = new Map<string, CollectableListing>();
  issuedCollectables.forEach((item) => {
    if (byID.has(item.id)) {
      let entry = byID.get(item.id)!;
      entry.instances.push(item);

      var d = new Date(item.issued_at);
      if (d < entry.first_acquired) {
        entry.first_acquired = d;
      }
    } else {
      byID.set(item.id, {
        ...item,
        instances: [item],
        first_acquired: new Date(item.issued_at),
      });
    }
  });

  return Array.from(byID.values())
}

type UserCollectionListProps = {
  collection: IssuedCollectable[];
}

type CollectableListing = Collectable & {
  instances: IssuedCollectable[];
  first_acquired: Date;
}

type RarityFilters = {
  [key in Rarity]?: boolean;
}

type Filters = RarityFilters & {
  first_edition?: boolean;
  subscriber?: boolean;
  verified?: boolean;
}

const filterListing = (filters: Filters, listing: IssuedCollectable): boolean => {
  if (filters.first_edition && filters.first_edition !== (listing.edition == "First Edition")) {
    return false;
  }
  if (filters.subscriber && filters.subscriber !== listing.subscriber) {
    return false;
  }
  if (filters.verified && filters.verified !== listing.verified) {
    return false;
  }

  // Should compute this once before filtering
  let rarity = listing.rarity as Rarity;
  if (!filters[rarity]) {
    return false;
  }

  return true;
}

export const UserCollectionList: Component<UserCollectionListProps> = (props) => {
  const [filters, setFilters] = createSignal<Filters>({
    [Rarity.Common]: true,
    [Rarity.Uncommon]: true,
    [Rarity.Rare]: true,
    [Rarity.SuperRare]: true,
    [Rarity.UltraRare]: true,
  });
  const [filter, setFilter] = createSignal("")

  const textFiltered = () => {
    return props.collection.filter((item) => {
      if (filter() === "") {
        return true;
      }
      let f = filter().toLowerCase();
      return item.name.toLowerCase().includes(f) || item.author.name.toLowerCase().includes(f);
    });
  }

  const rarityCounts = () => {
    let counts = {
      [Rarity.Common]: 0,
      [Rarity.Uncommon]: 0,
      [Rarity.Rare]: 0,
      [Rarity.SuperRare]: 0,
      [Rarity.UltraRare]: 0,
    };
    textFiltered().forEach((item) => {
      counts[item.rarity as Rarity]++;
    });
    return counts;
  }

  const filteredCollection = () => {
    return textFiltered().filter((item) => {
      if (!filterListing(filters(), item)) {
        return false;
      }
      return true;
    });
  }

  const listings = () => {
    return ListingFromCollectables(filteredCollection());
  }


  return (
    <div class={styles.UserListing} >
      <section class={styles.Filters}>
        <Stack direction="column" spacing={1}>
          <TextField label="Search" variant="outlined" autoComplete="off" onChange={(e) => { setFilter(e.target.value); }} />

          <Stack direction="row" spacing={1}>
            <Chip
              icon={<Check />}
              onClick={() => { setFilters((f) => ({ ...f, first_edition: !f.first_edition })) }}
              color={filters().first_edition ? "primary" : "default"}
              label="First Edition"
            />
            <Chip
              icon={<Favorite />}
              onClick={() => { setFilters((f) => ({ ...f, subscriber: !f.subscriber })) }}
              color={filters().subscriber ? "primary" : "default"}
              label="Subscriber"
            />
            <Chip
              icon={<Verified />}
              onClick={() => { setFilters((f) => ({ ...f, verified: !f.verified })) }}
              color={filters().verified ? "primary" : "default"}
              label="Verified"
            />
          </Stack>

          <Stack direction="row" spacing={1}>
            <For each={rarities}>
              {(rarity) => (
                <Chip
                  onClick={() => { setFilters((f) => ({ ...f, [rarity]: !filters()[rarity] })) }}
                  color={filters()[rarity] ? "primary" : "default"}
                  label={rarity + " (" + rarityCounts()[rarity] + ")"}
                />
              )}
            </For>
          </Stack>
        </Stack>
      </section>
      <section class={styles.Stats}>
        <Stack direction="row" justifyContent='space-between'>
          <div class={styles.Stat}>
            <div class={styles.StatLabel}>Total Knives:</div>
            <div>{props.collection.length}</div>
          </div>
          <div class={styles.Stat}>
            <div class={styles.StatLabel}>Matching Knives:</div>
            <div>{filteredCollection().length}</div>
          </div>
          <div class={styles.Stat}>
            <div class={styles.StatLabel}>Unique Knives:</div>
            <div>{listings().length}</div>
          </div>
        </Stack>
      </section>
      <section>
        <For each={listings()} >
          {(item) => (
            <UserCollectionListing listing={item} />
          )}
        </For>
      </section>
    </div>
  )
}

type UserCollectionListingProps = {
  listing: CollectableListing;
}

const UserCollectionListing: Component<UserCollectionListingProps> = (props) => {
  let clsList = {
    [styles.Listing]: true,
    [styles[props.listing.rarity.replace(/\s/g, "")]]: true,
  };

  return (
    <>
      <div classList={clsList}>
        <InlineImage image_url={props.listing.image_url} rarity={props.listing.rarity} />
        <div class={styles.ListingBody}>
          <div class={styles.ListingHeader}>
            <div class={styles.ListingTitle}>
              <div class={styles.ListingName}> {props.listing.name}</div>
              <div class={styles.Author}>{props.listing.author.name}</div>
            </div>
            <div class={styles.ListingIssued}>
              <div class={styles.Heading}>First Earned</div>
              <div class={styles.Val}>{props.listing.first_acquired.toLocaleDateString()}</div>
            </div>
          </div>
          <For each={props.listing.instances}>
            {(instance) => (
              <A href={`/knife/${instance.instance_id}`}>
                <div class={styles.ListingInstance}>
                  <div class={styles.InstanceDate}>{new Date(instance.issued_at).toLocaleDateString()}</div>
                  <div class={styles.InstanceBadges}>
                    <Show when={instance.verified}>
                      <Verified fontSize="small" titleAccess="Verified Issue" />
                    </Show>
                    <Show when={instance.subscriber}>
                      <Favorite fontSize="small" titleAccess="Verified Issue" />
                    </Show>
                    <Show when={instance.edition == "First Edition"}>
                      <Check fontSize="small" titleAccess="Verified Issue" />
                    </Show>
                  </div>
                </div>
              </A>
            )}
          </For>
        </div>
      </div>
    </>
  )
}

type InlineImageProps = {
  image_url: string;
  rarity: Rarity;
}

const InlineImage = (props: InlineImageProps) => {
  return (
    <div class={styles.InlineImage}>
      <svg class={styles.SVG} viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">
        <polygon points="50,0 75,25 75,75 50,100 25,75 25,25" fill="none" />
      </svg>
      <img src={props.image_url} />
    </div>
  );
}
