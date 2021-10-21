#define FUSE_USE_VERSION  26

#include <fuse.h>
#include <string.h>
#include <errno.h>
#include <fcntl.h>
#include <stdlib.h>

const char *hello_contents = "hello, world!\n";
const char *hello_path = "/hello";

static int my_getattr(const char *path, struct stat *stats) {
    int res = 0;
    memset(stats, 0, sizeof(struct stat));
    if (strcmp(path, "/") == 0) {
        stats->st_mode = S_IFDIR | 0755;
        stats->st_nlink = 2;
    } else if (strcmp(path, hello_path) == 0) {
        stats->st_mode = S_IFREG | 0444;
        stats->st_nlink = 1;
        stats->st_size = (off_t) strlen(hello_contents);
    } else
        res = -ENOENT;

    return res;
}

static int my_readdir(const char *path, void *buf, fuse_fill_dir_t fill_dir,
                      off_t offset, struct fuse_file_info *fi) {
    (void) offset;
    (void) fi;

    if (strcmp(path, "/") != 0)
        return -ENOENT;
    fill_dir(buf, ".", NULL, 0);
    fill_dir(buf, "..", NULL, 0);
    fill_dir(buf, hello_path + 1, NULL, 0);
    return 0;
}

static int my_open(const char *path, struct fuse_file_info *fi) {
    if (strcmp(path, hello_path) != 0) {
        return -ENOENT;
    }
    else if ((fi->flags & 3) != O_RDONLY) {
        return -EACCES;
    }
    return 0;
}

static int my_read(const char *path, char *buf, size_t size,
                   off_t offset, struct fuse_file_info *fi) {
    size_t len;
    (void) fi;

    if (strcmp(path, hello_path) != 0) {
        return -ENOENT;
    }

    len = strlen(hello_contents);
    if (offset < len) {
        if (offset + size > len) {
            size = len - offset;
        }
        memcpy(buf, hello_contents + offset, size);
    } else {
        size = 0;
    }

    return (int) size;
}

static struct fuse_operations f_operations = {
        .getattr = my_getattr,
        .readdir = my_readdir,
        .open = my_open,
        .read = my_read,
};

int main(int argc, char *argv[]) {
    return fuse_main(argc, argv, &f_operations, NULL);
}