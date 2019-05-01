#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
#include <sys/types.h>
#include <math.h>
#include<signal.h>
#define NUM_THREADS     sysconf(_SC_NPROCESSORS_ONLN)

void handleSignal(int sig) 
{ 
  FILE *fp;

  char file_name[50];
  sprintf(file_name, "/proc/%d/schedstat", getpid());
  fp = fopen(file_name, "r"); // read mode
 
  if (fp == NULL)
  {
    exit(1);
  }

  char ch;
  while((ch = fgetc(fp)) != EOF){
    printf("%c", ch);
  }
  printf("::::::::::\n"); // Seperator for parsing
  sprintf(file_name, "/proc/%d/stat", getpid());
  fp = fopen(file_name, "r"); // read mode
 
  if (fp == NULL)
  {
    exit(1);
  }

  while((ch = fgetc(fp)) != EOF){
    printf("%c", ch);
  }

  printf("::::::::::\n");
  sprintf(file_name, "/proc/%d/sched", getpid());
  fp = fopen(file_name, "r"); // read mode
 
  if (fp == NULL)
  {
    exit(1);
  }

  while((ch = fgetc(fp)) != EOF){
    printf("%c", ch);
  }
  exit(0);
} 

void *IOIntensive(void* nothing)
{
  FILE *file;
  file = tmpfile();
  char dummyData[] = "deadbeef";
  for (int i = 0; i < 100; i++) {
    int n = fwrite(&dummyData, 1, 8, file);
    if (n != 8) {
      printf("Uh oh, only wrote %d\n", n);
    }
    fflush(file);
    fseek(file, -3, SEEK_CUR);
    fseek(file, 4096 * 1024 * 5, SEEK_CUR);
  }
  fclose(file);
  return nothing;
}

void *CPUIntensive(void* nothing)
{
  double pi = 0.0;
  uint counter = 0;
  while (1) {
    counter++;
    pi += 4.0 * pow(-1, (double)counter) / (double)((2*counter)+1);
  }
  return nothing;
}

void joinThreads(pthread_t threads[]) {
  long t;
  for(t=0; t<NUM_THREADS; t++){
    pthread_join(threads[t], NULL);
  }
}

int main (int argc, char *argv[])
{
  pthread_t threads[NUM_THREADS];
  int rc;
  long t;
  if (argc != 2) {
    printf("Must provide two arguments\n");
    return 1;
  }
  signal(SIGQUIT, handleSignal);
  if (strncmp(argv[1], "io", 15) == 0) {
    IOIntensive((void *) t);
  } else if (strncmp(argv[1], "cpu", 15) == 0) {
    CPUIntensive((void *) t);
  } else if (strncmp(argv[1], "mixed-threaded", 15) == 0) {
    for(t=0; t<NUM_THREADS; t++){
      if (t % 2 == 0) {
        rc = pthread_create(&threads[t], NULL, IOIntensive, (void *) t);  
      } else {
        rc = pthread_create(&threads[t], NULL, CPUIntensive, (void *) t);
      }
      if (rc){
        printf("ERROR; return code from pthread_create() is %d\n", rc);
        return 1;
      }
    }
    joinThreads(threads);
  } else if (strncmp(argv[1], "io-threaded", 15) == 0) {
    for(t=0; t<NUM_THREADS; t++){
      rc = pthread_create(&threads[t], NULL, IOIntensive, (void *) t);
      if (rc){
        printf("ERROR; return code from pthread_create() is %d\n", rc);
        return 1;
      }
    }
    joinThreads(threads);
  } else if (strncmp(argv[1], "cpu-threaded", 15) == 0) {
    for(t=0; t<NUM_THREADS; t++){
      rc = pthread_create(&threads[t], NULL, CPUIntensive, (void *) t);
      if (rc){
        printf("ERROR; return code from pthread_create() is %d\n", rc);
        return 1;
      }
    }
    joinThreads(threads);
  } else {
    for(t=0; t<NUM_THREADS; t++){
      rc = pthread_create(&threads[t], NULL, IOIntensive, (void *) t);
      if (rc){
        printf("ERROR; return code from pthread_create() is %d\n", rc);
        return 1;
      }
    }
    joinThreads(threads);
  }

  // /* Last thing that main() should do */
  pthread_exit(NULL);
  return 0;
}
