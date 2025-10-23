import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { provideHttpClient, withFetch } from '@angular/common/http';

import { AppComponent } from './app.component';

@NgModule({
  declarations: [],
  imports: [
    BrowserModule,
    AppComponent
  ],
  providers: [provideHttpClient(withFetch())],
  bootstrap: []
})
export class AppModule { }