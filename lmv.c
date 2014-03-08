#include <stdio.h>
#include <stdlib.h>
#include <errno.h>
#include <string.h>
#include <sys/stat.h>
#include <openssl/sha.h>
#include <curl/curl.h>
#include <stdbool.h>
#include <libgen.h>

struct Response {
    char *body;
    size_t size;
};

static char* WriteMemoryCallback(void *contents, size_t size, size_t nmemb, void *userp) {

    size_t realsize = size * nmemb;
    struct Response *resp = (struct Response *)userp;
 
    resp->body = realloc(resp->body, resp->size + realsize + 1);

    if(resp->body == NULL) {
  
        printf("not enough memory (realloc returned NULL)\n");
        return 0;

    }
 
    memcpy(&(resp->body[resp->size]), contents, realsize);
    resp->size += realsize;
    resp->body[resp->size] = 0;
 
    return resp->body;

}

bool file_is_accessible (const char * file_path) {

    FILE *f = fopen(file_path, "r");

    if (f) {
        fclose(f);
        return true;
    }

    return false;

}

char* name (char * file_path) {

    return basename(file_path);

}

static unsigned int size (const char * file_path) {
    
    struct stat sb;
    
    if (stat(file_path, & sb) != 0) {

        fprintf (stderr, "'stat' failed for '%s': %s.\n",
                 file_path, strerror (errno));
        
        exit (EXIT_FAILURE);
    
    }
    
    return sb.st_size;

}

char* hash (const char * file_path) {

    unsigned char data[size (file_path)];
    FILE *fp;

    fp = fopen(file_path, "rb");
    fread(&data, 1, size (file_path), fp);
    fclose(fp);

    unsigned char hash[SHA512_DIGEST_LENGTH];

    SHA512(data, sizeof(data), hash);

    char* digest = calloc(SHA512_DIGEST_LENGTH*2+1, sizeof(char));

    for (int i = 0; i < SHA512_DIGEST_LENGTH; i++) {

        sprintf(&digest[i*2], "%02x", (unsigned int)hash[i]);
    
    }

    return digest;

}

void down (char* token, char* server) {

    CURL *curl_handle;
    CURLcode res;
    char *tuplestring;
    char *brkb;
    struct Response chunk;
    char *target;

    asprintf(&target, "%s/%s", server, token);

    chunk.body = malloc(1);  /* will be grown as needed by the realloc above */ 
    chunk.size = 0;    /* no data at this point */ 
 
    curl_global_init(CURL_GLOBAL_ALL);
 
    curl_handle = curl_easy_init();
 
    curl_easy_setopt(curl_handle, CURLOPT_URL, target);
 
    curl_easy_setopt(curl_handle, CURLOPT_WRITEFUNCTION, WriteMemoryCallback);
 
    curl_easy_setopt(curl_handle, CURLOPT_WRITEDATA, (void *)&chunk);
    
    res = curl_easy_perform(curl_handle);
 
    if(res != CURLE_OK) {
        fprintf(stderr, "curl_easy_perform() failed: %s\n",
            curl_easy_strerror(res));
    }

    int i=0;

    char *parts[3];

    for (tuplestring = strtok_r(chunk.body, ";", &brkb); 
            tuplestring;
            tuplestring = strtok_r(NULL, ";", &brkb)) {

        parts[i] = tuplestring;
        
        i++;
    }
 
    int byte_count = 1;

    char data[atoi(parts[2])];

    FILE *urandom;
    FILE *output;

    int k = 1;

    while (strcmp(parts[0], hash(parts[1])) != 0) {

        printf("%d: %s\n", k, hash(parts[1]));

        urandom = fopen("/dev/urandom", "rb");

        fread(&data, byte_count, atoi(parts[2]), urandom);
        fclose(urandom);

        FILE *create = fopen(parts[1], "w+");
        fclose(create);

        output = fopen(parts[1], "wb+");

        fwrite(&data, byte_count, atoi(parts[2]), output);
        fclose(output);

        k++;

    }

    printf("Download complete");

}

void up (char *hash, unsigned int size, char *name, char *server) {

    char sizestr[50]; 
    CURL *curl;
    CURLcode res;

    struct Response chunk;

    chunk.body = malloc(1);  /* will be grown as needed by the realloc above */ 
    chunk.size = 0;    /* no data at this point */     
 
    struct curl_httppost *formpost=NULL;
    struct curl_httppost *lastptr=NULL;
    struct curl_slist *headerlist=NULL;

    curl_formadd(&formpost,
                 &lastptr,
                 CURLFORM_COPYNAME, "hash",
                 CURLFORM_CONTENTSLENGTH, SHA512_DIGEST_LENGTH*2,
                 CURLFORM_COPYCONTENTS, hash,
                 CURLFORM_END);

    snprintf(sizestr, 50, "%d", size); 

    curl_formadd(&formpost,
                 &lastptr,
                 CURLFORM_COPYNAME, "size",
                 CURLFORM_COPYCONTENTS, sizestr,
                 CURLFORM_END);

    curl_formadd(&formpost,
                 &lastptr,
                 CURLFORM_COPYNAME, "name",
                 CURLFORM_COPYCONTENTS, name,
                 CURLFORM_END);

    curl = curl_easy_init();
    
    if (curl) {
        
        curl_easy_setopt(curl, CURLOPT_URL, server);
        curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, WriteMemoryCallback);
        curl_easy_setopt(curl, CURLOPT_WRITEDATA, (void *)&chunk);
        curl_easy_setopt(curl, CURLOPT_HTTPPOST, formpost);
        
        res = curl_easy_perform(curl);
        
        if(res != CURLE_OK) {
            
            /*fprintf(stderr, "curl perform failed: %s\n", curl_easy_strerror(res));*/
        
        }

        printf("\n\'%s\' ==> \'%s\'\n", name, chunk.body);

        curl_easy_cleanup(curl);
        curl_formfree(formpost);
     
    }
 
}

int main (int argc, char ** argv) {

    char *server_address = "127.0.0.1:8081"; /* Set this to your register address */

    char *upload_address = NULL;
    char *download_address = NULL;

    asprintf(&upload_address, "%s/upload", server_address);
    asprintf(&download_address, "%s/download", server_address);

    curl_global_init(CURL_GLOBAL_ALL);

    if (argc < 2) {

        printf("\nLudicrous MV v0.1\n");

        printf("\nUsage:\n\n");
        printf("\tlmv <path>\tupload a file\n");
        printf("\tlmv <token>\tdownload a file\n\n");

        return 0;

    }
 
    for (int i = 1; i < argc; i++) {

    	/* Upload the file */

        if (file_is_accessible(argv[i])) {

            printf("Uploading: %s", name(argv[i])); 

            up(hash(argv[i]), size(argv[i]), name(argv[i]), upload_address);
            printf("\n");

        }
        
        /* Attempt to download a file */

        else {

            down(argv[i], download_address);

        }

    }
   
    return 0;

}
