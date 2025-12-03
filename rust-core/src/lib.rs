use libc::{c_int, size_t, c_uchar};
use oqs::kem::{Kem, Algorithm};
use std::slice;
use std::ptr;

// Constantes para o Go saber o tamanho dos buffers
// Kyber-768 (ML-KEM-768) é o padrão NIST Level 3
const KEM_ALG: Algorithm = Algorithm::Kyber768;

/// Retorna o tamanho da Chave Pública
#[unsafe(no_mangle)]
pub extern "C" fn get_public_key_len() -> size_t {
    let kem = Kem::new(KEM_ALG).unwrap();
    kem.length_public_key()
}

/// Retorna o tamanho da Chave Privada
#[unsafe(no_mangle)]
pub extern "C" fn get_secret_key_len() -> size_t {
    let kem = Kem::new(KEM_ALG).unwrap();
    kem.length_secret_key()
}

/// Retorna o tamanho do Texto Cifrado (Ciphertext)
#[unsafe(no_mangle)]
pub extern "C" fn get_ciphertext_len() -> size_t {
    let kem = Kem::new(KEM_ALG).unwrap();
    kem.length_ciphertext()
}

/// Retorna o tamanho do Segredo Compartilhado
#[unsafe(no_mangle)]
pub extern "C" fn get_shared_secret_len() -> size_t {
    let kem = Kem::new(KEM_ALG).unwrap();
    kem.length_shared_secret()
}

/// Gera um par de chaves (Pública e Privada)
#[unsafe(no_mangle)]
pub unsafe extern "C" fn pqc_keypair(
    public_key_ptr: *mut c_uchar,
    secret_key_ptr: *mut c_uchar
) -> c_int {
    let kem = match Kem::new(KEM_ALG) {
        Ok(k) => k,
        Err(_) => return -1,
    };

    let (pk, sk) = match kem.keypair() {
        Ok(pair) => pair,
        Err(_) => return -2,
    };

    // Bloco unsafe explícito para operações de memória
    unsafe {
        ptr::copy_nonoverlapping(pk.as_ref().as_ptr(), public_key_ptr, pk.len());
        ptr::copy_nonoverlapping(sk.as_ref().as_ptr(), secret_key_ptr, sk.len());
    }

    0
}

/// Encapsula (Cria segredo compartilhado)
#[unsafe(no_mangle)]
pub unsafe extern "C" fn pqc_encaps(
    public_key_ptr: *const c_uchar,
    ciphertext_out: *mut c_uchar,
    shared_secret_out: *mut c_uchar
) -> c_int {
    let kem = match Kem::new(KEM_ALG) {
        Ok(k) => k,
        Err(_) => return -1,
    };

    let pk_len = kem.length_public_key();
    
    // Recupera a chave pública do ponteiro
    let public_key = unsafe {
        let pk_slice = slice::from_raw_parts(public_key_ptr, pk_len);
        kem.public_key_from_bytes(pk_slice).unwrap()
    };

    let (ct, ss) = match kem.encapsulate(&public_key) {
        Ok(res) => res,
        Err(_) => return -2,
    };

    unsafe {
        ptr::copy_nonoverlapping(ct.as_ref().as_ptr(), ciphertext_out, ct.len());
        ptr::copy_nonoverlapping(ss.as_ref().as_ptr(), shared_secret_out, ss.len());
    }

    0
}

/// Decapsula (Recupera segredo com chave privada)
#[unsafe(no_mangle)]
pub unsafe extern "C" fn pqc_decaps(
    ciphertext_ptr: *const c_uchar,
    secret_key_ptr: *const c_uchar,
    shared_secret_out: *mut c_uchar
) -> c_int {
    let kem = match Kem::new(KEM_ALG) {
        Ok(k) => k,
        Err(_) => return -1,
    };

    let sk_len = kem.length_secret_key();
    let ct_len = kem.length_ciphertext();

    // Recupera chaves e ciphertext dos ponteiros
    let (secret_key, ciphertext) = unsafe {
        let sk_slice = slice::from_raw_parts(secret_key_ptr, sk_len);
        let ct_slice = slice::from_raw_parts(ciphertext_ptr, ct_len);
        
        (
            kem.secret_key_from_bytes(sk_slice).unwrap(),
            kem.ciphertext_from_bytes(ct_slice).unwrap()
        )
    };

    let ss = match kem.decapsulate(&secret_key, &ciphertext) {
        Ok(s) => s,
        Err(_) => return -2,
    };

    unsafe {
        ptr::copy_nonoverlapping(ss.as_ref().as_ptr(), shared_secret_out, ss.len());
    }

    0
}