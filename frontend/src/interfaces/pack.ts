

export interface Pack {
  id: number;
  slug: string;
  description: string;
  created_at: string;
  updated_at: string;
  is_public: boolean;
}

export interface Packs {
  packs: Pack[];
}
