#include <stdio.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>
#include <stdlib.h>

#include <sys/prctl.h>

int main(int argc, char *argv[]) {
    int pid = getpid();

    // Finding out length of first command line argument - from cmdline file
    int buff_cap = 1024;

    char filename[buff_cap];
    sprintf(filename, "/proc/%d/cmdline", pid);

    char buff[buff_cap];
    FILE* f = fopen(filename, "r");
    if (f == NULL) {
        printf("failed to open cmdline file");
        return 1;
    }
    char* tmp = fgets(buff, buff_cap, f);
    if (tmp == NULL) {
        printf("failed to read from cmdline file");
        return 1;
    }
    if (fclose(f) == -1) {
        printf("failed to close cmdline file");
        return 1;
    }

    size_t cmd_len = strlen(buff);

    // Then finding out pointer to start of command line arguments
    sprintf(filename, "/proc/%d/stat", pid);
    FILE* stats = fopen(filename, "r");
    if (stats == NULL) {
        printf("failed to open stats file");
        return 1;
    }

    tmp = fgets(buff, buff_cap, stats);
    if (!tmp) {
        printf("failed to read from stats file");
        return 1;
    }
    if (fclose(stats) == -1) {
        printf("failed to close stats file");
        return 1;
    }

    //searching for argv_start
    tmp = strchr(buff, ' ');
    for (int i = 0; i < 46; i++) {
        if (!tmp) {
            printf("format of stat file is not correct");
            return 1;
        }
        tmp = strchr(tmp+1, ' ');
    }
    if (!tmp) {
        printf("format of stat file is not correct");
        return 1;
    }

    unsigned long argv_start;
    int cnt = sscanf(tmp, "%lu", &argv_start);
    if (cnt != 1) {
        printf("format of stat file is not correct");
        return 1;
    }

    // Changing command line!
    if (prctl(PR_SET_MM, PR_SET_MM_ARG_START, argv_start + cmd_len + 1, 0, 0) == -1) {
        printf("Prctl did not work. Maybe permissions are needed?");
        return 1;
    };

    // We need PID of process and some time to check if it has worked
    printf("Process PID: %d\n", pid);
    sleep(300);
    return 0;
}
