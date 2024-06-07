import { ComponentFixture, TestBed } from '@angular/core/testing';

import { EnvironmentDetailsComponent } from './environment-details.component';

describe('EnvironmentDetailsComponent', () => {
  let component: EnvironmentDetailsComponent;
  let fixture: ComponentFixture<EnvironmentDetailsComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [EnvironmentDetailsComponent]
    })
    .compileComponents();
    
    fixture = TestBed.createComponent(EnvironmentDetailsComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
