// 与后端 model/dto 对应的前端类型定义
export type Role = 'user' | 'admin';
export type UserStatus = 'active' | 'banned';
export type PostType = 'image' | 'video';
export type PostStatus = 'pending' | 'approved' | 'rejected';
export type MeetupStatus = 'open' | 'full' | 'cancelled' | 'completed';
export type MeetupJoinStatus = 'joined' | 'cancelled';
export type InteractionType = 'like' | 'favorite' | 'forward';
export type PetSpecies = 'dog' | 'cat' | 'other';
export type PetGender = 'male' | 'female';
export type PetStatus = 'active' | 'hidden';
export type ChatMessageType = 'text' | 'image' | 'sticker';
export type TopicStatus = 'active' | 'archived';

export interface User {
  id: string;
  username: string;
  nickname: string;
  avatar: string;
  bio: string;
  city: string;
  role: Role;
  status: UserStatus;
  follow_count: number;
  follower_count: number;
  created_at: string;
}

export interface TokenResponse {
  token: string;
  user: User;
}

export interface Pet {
  id: string;
  owner_id: string;
  name: string;
  species: PetSpecies;
  breed: string;
  birthday: string;
  gender: PetGender;
  personality: string;
  bio: string;
  avatar: string;
  album: string[];
  city: string;
  status: PetStatus;
  follower_count: number;
  created_at: string;
}

export interface MediaItem {
  type: 'image' | 'video';
  url: string;
}

export interface Post {
  id: string;
  author_id: string;
  author_name: string;
  author_avatar: string;
  pet_ids: string[];
  content: string;
  type: PostType;
  media: MediaItem[];
  topics: string[];
  location: string;
  city: string;
  status: PostStatus;
  reject_reason: string;
  like_count: number;
  comment_count: number;
  favorite_count: number;
  forward_count: number;
  created_at: string;
  liked: boolean;
  favorited: boolean;
}

export interface Comment {
  id: string;
  post_id: string;
  author_id: string;
  author_name: string;
  content: string;
  created_at: string;
}

export interface Topic {
  id: string;
  name: string;
  description: string;
  post_count: number;
  status: TopicStatus;
  created_at: string;
}

export interface MeetupParticipant {
  user_id: string;
  username: string;
  joined_at: string;
  status: MeetupJoinStatus;
}

export interface Meetup {
  id: string;
  creator_id: string;
  creator_name: string;
  title: string;
  description: string;
  city: string;
  location: string;
  meet_time: string;
  max_people: number;
  status: MeetupStatus;
  joined_count: number;
  joined: boolean;
  participants: MeetupParticipant[];
  created_at: string;
}

export interface ChatMessage {
  id: string;
  from_id: string;
  to_id: string;
  type: ChatMessageType;
  content: string;
  read: boolean;
  created_at: string;
}

export interface Conversation {
  peer_id: string;
  peer_name: string;
  peer_avatar: string;
  last_message: string;
  last_time: string;
  unread_count: number;
}

export interface AuditLog {
  id: string;
  user_id: string;
  username: string;
  action: string;
  resource: string;
  resource_id: string;
  detail: string;
  ip: string;
  created_at: string;
}

export interface PageData<T> {
  items: T[];
  total: number;
  page: number;
  page_size: number;
}

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}
