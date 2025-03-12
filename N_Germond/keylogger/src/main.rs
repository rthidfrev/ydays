use std::process::Command;
use device_query::{DeviceEvents, DeviceEventsHandler, Keycode, MouseButton, MousePosition};
use keycode::{KeyboardState, KeyMap, KeyMappingId, KeyState};
use std::sync::{Arc,Mutex};
use std::time::Duration;
use std::thread;


fn main() {
    let mut caps_lock: Arc<Mutex<bool>> = Arc::new(Mutex::new(false));

    let event_handler = DeviceEventsHandler::new(Duration::from_millis(10))
        .expect("Could not initialize event loop");

    let _key_press_guard_0 = event_handler.on_key_down(move |keycode: &Keycode| {
        match keycode {
            Keycode::LShift => {
                let mut value = caps_lock.lock().unwrap() ;
                *value = true;
                println!("{}", *value);
                println!("La touche 'shift' a été pressée !");
            },
            _ => {}
        }
    });

    let _key_press_guard_1 = event_handler.on_key_up(move |keycode: &Keycode| {
      match keycode {
          Keycode::LShift => {
              let mut value = caps_lock.lock().unwrap() ;
              *value = false;
              println!("{}", *value);
              println!("La touche 'shift' a été relaché !");
          },
          _ => {}
      }
  });

    //let maj: bool = keyboard_state.update_key(shift, KeyState::Pressed);
    //let _key_press_guard = event_handler.on_key_down(|keycode: &Keycode| {
    //  println!("{}", maj);
    //});

    loop {
        thread::sleep(Duration::from_secs(1000));
    }
}
