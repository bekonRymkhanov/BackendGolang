import {Component, OnInit} from '@angular/core';
import {User} from "../models";
import {OneXBetService} from "../one-xbet.service";
import {AuthService} from "../auth.service";

import {NgIf} from "@angular/common";

@Component({
  selector: 'app-user',
  standalone: true,
  imports: [
    NgIf
  ],
  templateUrl: './user.component.html',
  styleUrl: './user.component.css'
})
export class UserComponent implements OnInit{
  loaded:boolean=false
  newUser:User;

  ngOnInit(): void {
    this.authservice.currentUser$.subscribe(user => {
      if (!user) {
        return;
      }
      this.newUser = user;
      this.loaded=true
    }
    )
  }
  constructor(private httpService:OneXBetService,private authservice:AuthService) {
    this.newUser={
      "id":0,
      "name":'',
      "email":'',
      "password":'',
      "is_admin":false,
      "created_at":"",
      activated:false,
      favorite_episodes:[],
    }
  }
}
