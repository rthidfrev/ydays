#[no_mangle]

pub extern "C" fn _start() -> ! {
    unsafe { lib::_exit(0) }
}

