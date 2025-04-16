export interface Book {
  id: number;
  title: string;
  author: string;
  main_genre: string;
  sub_genre: string;
  type: string;
  price: string;
  rating: number;
  people_rated: number;
  url: string;
  version: number;
}

export interface Genre {
  id: number;
  title: string;
  subgenre_count: number;
  url: string;
  version: number;
}

export interface SubGenre {
  id: number;
  title: string;
  main_genre: string;
  book_count: number;
  url: string;
  version: number;
}

export interface Comment {
  id: number;
  book_id: number;
  user_id: number;
  content: string;
  created_at: Date;
  version: number;
}

export interface Rating {
  id: number;
  book_id: number;
  user_id: number;
  score: number;
  created_at: Date;
  version: number;
}

export interface Metadata {
  CurrentPage: number;
  PageSize: number;
  FirstPage: number;
  LastPage: number;
  TotalRecords: number;
}

export interface BookFilters {
  title?: string;
  author?: string;
  main_genre?: string;
  sub_genre?: string;
  type?: string;
  page?: number;
  page_size?: number;
  sort?: string;
}

export interface FavoriteBook {
  id: number;
  user_id: number;
  book_name: string;
  created_at: string;
  is_admin: boolean;
}

export interface User {
  id?: number;
  name: string;
  email: string;
  password?: string;
  is_admin: boolean;
  activated?: boolean;
  created_at?: string;
  favorite_episodes?: any;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface ActivationResponse {
  user: User;
}

export interface AuthenticationResponse {
  authentication_token: {
    token: string;
    expiry: string;
  }
  user:User
}

export interface requestBookDetail {
  book: Book
}

