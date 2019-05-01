#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <string.h>
#include <sys/types.h>
#include <math.h>
#include<signal.h>
#define NUM_THREADS     sysconf(_SC_NPROCESSORS_ONLN)
#define SEPERATOR "::::::::::\n"
#define OFFSET 20000


FILE *filePtr;
pthread_mutex_t lock;

void printSchedulerInfo() 
{ 
  FILE *fp;

  char file_name[50];
  sprintf(file_name, "/proc/%d/schedstat", getpid());
  fp = fopen(file_name, "r"); // read mode
 
  if (fp == NULL)
  {
    printf("Foo\n");
    exit(1);
  }

  char ch;
  while((ch = fgetc(fp)) != EOF){
    printf("%c", ch);
  }
  printf(SEPERATOR); // Seperator for parsing
  sprintf(file_name, "/proc/%d/stat", getpid());
  fp = fopen(file_name, "r"); // read mode
 
  if (fp == NULL)
  {
    printf("Foo\n");
    exit(1);
  }

  while((ch = fgetc(fp)) != EOF){
    printf("%c", ch);
  }

  printf(SEPERATOR);
  sprintf(file_name, "/proc/%d/sched", getpid());
  fp = fopen(file_name, "r"); // read mode
 
  if (fp == NULL)
  {
    printf("Foo\n");
    exit(1);
  }

  while((ch = fgetc(fp)) != EOF){
    printf("%c", ch);
  }
  exit(0);
}


  uint counter;

void *writePiToDisk(void* counter)
{
  double pi = 0.0;
  uint start = 0;
  long realCounter = (long)&counter;
  while (start < OFFSET) {
    realCounter++;
    start++;
    pi += 4.0 * pow(-1, (double)realCounter) / (double)((2*realCounter)+1);
  }
  pthread_mutex_lock(&lock);
  fprintf(filePtr, "%.g\n", pi);
  pthread_mutex_unlock(&lock);
  return NULL;
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
  int i;
  long t;
  if (pthread_mutex_init(&lock, NULL) != 0) {
    printf("Failed to initialize lock\n");
    return 1;
  }
  if (argc != 2) {
    printf("Must provide two arguments\n");
    return 1;
  }
  filePtr = tmpfile();
  int withThreads = strncmp(argv[1], "serial", 15);

  for (i = 0; i < OFFSET; i ++) {
    for(t=0; t<NUM_THREADS; t++){
      if (withThreads != 0) {
        rc = pthread_create(&threads[t], NULL, writePiToDisk, (void *)((OFFSET * i) * t));
        if (rc){
          printf("ERROR; return code from pthread_create() is %d\n", rc);
          return 1;
        }  
      } else {
        writePiToDisk((void *)((OFFSET * i) * t));
      }
    }
    if (withThreads != 0) {
      joinThreads(threads);
    }
  }
  // /* Last thing that main() should do */
  pthread_exit(NULL);
  fflush(filePtr);
  printf("Printing scheduler info\n");
  printSchedulerInfo();
  return 0;
}
