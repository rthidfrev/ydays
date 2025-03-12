#![no_std]
#![no_main]

extern crate windows;
use windows::Win32::System::Memory::{VirtualAlloc, VirtualProtect, MEM_COMMIT, MEM_RESERVE, PAGE_READWRITE, PAGE_EXECUTE_READ};
use windows::Win32::Foundation::BOOL;
use core::ptr::null_mut;


#[no_mangle]

pub extern "C" fn _start() -> ! {
    unsafe {
        //Allouer 4096 octects (une page mémoire)
        let mem_size: usize = 0x1000;
        let allocated_mem = VirtualAlloc(
            null_mut(), // Laisser windows choisir l'adresse
            mem_size,
            MEM_COMMIT | MEM_RESERVE, // Réserver et allouer
            PAGE_EXECUTE_READWRITE, // Permission d'exécuter du code injecté
        );

        if allocated_mem.is_null() {
            panic!("Échec de l'allocation de mémoire !");
        }

        //Modification des permissions
        let mut old_protect: u32 = 0;
        let success: BOOL = VirtualProtect(
            allocated_mem,
            mem_size,
            PAGE_EXECUTE_READ, // On change pour exécution
            &mut old_protect,
        );

        if success.0 == 0 {
            panic!("Échec du changement de permissions !");
        } else {
            println!("Permissions mémoire modifiées avec succès !");

        // Affichage pour le debug
        core::arch::asm!("nop"); // Juste pour éviter que le compilateur optimise tout

        loop {} // Évite que le programme se termine

    }

}