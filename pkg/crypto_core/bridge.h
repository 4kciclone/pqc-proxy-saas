#ifndef PQC_BRIDGE_H
#define PQC_BRIDGE_H

#include <stdlib.h>
#include <stdint.h>

// Retorna tamanhos para alocação de memória
size_t get_public_key_len();
size_t get_secret_key_len();
size_t get_ciphertext_len();
size_t get_shared_secret_len();

// Funções principais
int pqc_keypair(unsigned char *pk, unsigned char *sk);

int pqc_encaps(const unsigned char *pk, 
               unsigned char *ct, 
               unsigned char *ss);

int pqc_decaps(const unsigned char *ct, 
               const unsigned char *sk, 
               unsigned char *ss);

#endif