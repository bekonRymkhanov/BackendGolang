import {ApplicationConfig, Injectable} from "@angular/core";
import {HttpHandler, HttpInterceptor, HttpRequest, provideHttpClient} from "@angular/common/http";
import {provideRouter} from "@angular/router";
import {routes} from "./app.routes";

@Injectable()
export class AuthInterceptor implements HttpInterceptor {

  constructor() {
  }

  intercept(req: HttpRequest<any>, next: HttpHandler) {
    const access = localStorage.getItem("access");
    if (access) {
      const newReq = req.clone({
        headers: req.headers.set('Authorization', `Bearer ${access}`)
      });
      return next.handle(newReq);
    }

    return next.handle(req);
  }
}
