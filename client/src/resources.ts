export type User = {
  id: string;
  name: string;
};

export type Collectable = {
  id: string;
  name: string;
  author: User;
  rarity: Rarity;
  image_path: string;
};

export type IssuedCollectable = Collectable & {
  instance_id: string;
  owner: User;
  verified: boolean;
  subscriber: boolean;
  edition: string;
  issued_at: string;
  deleted: boolean;
};

export enum Rarity {
  Common = 'Common',
  Uncomomn = 'Uncommon',
  Rare = 'Rare',
  SuperRare = 'Super Rare',
  UltraRare = 'Ultra Rare',
};

export const rarityclass = (r: Rarity): string => {
  return r.toString().toLowerCase().replaceAll(' ', '-')
};
