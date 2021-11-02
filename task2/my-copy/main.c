#include <stdio.h>
#include <fcntl.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
#include <inttypes.h>
#include <sys/stat.h>
#include <liburing.h>

#define check_exit(res, msg) {\
    if (res < 0) {\
        fprintf(stderr, msg, strerror(-res));\
        return -1;\
    }\
}

#define check_cond(cond, msg) {\
    if (cond) {\
        fprintf(stderr, msg);\
        return -1;\
    }\
}

struct request_data {
    int read;
    size_t size, offset;
};

int submit_read(size_t read_size, size_t file_size, size_t *offset, struct iovec* iov,
                struct request_data* r_data, int id, int fd, struct io_uring* ring) {
    if (read_size > file_size - *offset) {
        read_size = file_size - *offset;
    }

    struct io_uring_sqe* sqe = io_uring_get_sqe(ring);
    check_cond(sqe == NULL, "Could not get SQE.\n")

    r_data[id].read = 1;
    r_data[id].size = read_size;
    r_data[id].offset = *offset;

    io_uring_prep_read_fixed(sqe, fd, iov[id].iov_base, read_size, *offset, id);
    sqe->user_data = (uint64_t) id;
    *offset += read_size;
    return 0;
}

int submit_write(struct iovec* iov, struct request_data* r_data,
                 int id, int fd, struct io_uring* ring) {
    struct io_uring_sqe* sqe = io_uring_get_sqe(ring);
    check_cond(sqe == NULL, "Could not get SQE.\n");

    r_data[id].read = 0;
    io_uring_prep_write_fixed(sqe, fd, iov[id].iov_base, r_data[id].size, r_data[id].offset, id);
    sqe->user_data = (uint64_t) id;
    return 0;
}

int main(int argc, char* argv[]) {
    int N = 4;
    size_t read_size = 256 * 1024;

    check_cond((argc != 3 && argc != 5), "Arguments were not provided");
    if (argc == 5) {
        N = strtol(argv[3], NULL, 10);
        read_size = strtol(argv[4], NULL, 10);
    }

    int input = open(argv[1], O_RDONLY);
    check_cond(input < 0, "Error while opening file\n");

    int output = open(argv[2], O_WRONLY | O_CREAT | O_TRUNC, 0755);
    check_cond(output < 0, "Error while opening file\n");

    struct io_uring ring;
    int res = io_uring_queue_init(2 * N, &ring, 0);
    check_exit(res, "Error while queue_init: %s\n");

    char* buffers = malloc(N * read_size);
    check_cond(buffers == NULL, "Error while allocating memory for buffer\n");

    struct iovec iov[N];
    for (int i = 0; i < N; ++i) {
        iov[i].iov_len = read_size;
        iov[i].iov_base = buffers + i * read_size;
    }
    res = io_uring_register_buffers(&ring, iov, N);
    check_exit(res, "Error registering buffers: %s\n");

    struct stat st;
    res = fstat(input, &st);
    check_exit(res, "Error while checking input file size: %s\n");
    check_cond(!S_ISREG(st.st_mode), "Wrong file type\n");

    size_t offset = 0;
    size_t to_write = st.st_size;
    size_t file_size = st.st_size;
    //printf("Size: %ld\n", st.st_size);

    struct request_data r_data[N];
    for (int i = 0; i < N && offset != file_size; ++i) {
        if (submit_read(read_size, file_size, &offset, iov, r_data, i,
                    input, &ring) < 0) {
            return -1;
        }
        io_uring_submit(&ring);
    }

    struct io_uring_cqe *cqe;
    bool can_peek = true;
    while (to_write > 0) {
        if (can_peek) {
            res = io_uring_peek_cqe(&ring, &cqe);
            if (res == -EAGAIN) {
                can_peek = false;
                continue;
            }
        } else {
            can_peek = true;
            res = io_uring_wait_cqe(&ring, &cqe);
        }

        check_exit(res, "Error waiting for completion: %s\n");
        //printf("!type_read: %d, read: %d, need: %zu, %lu, offs: %zu, to_write: %zu\n", r_data[(int)cqe->user_data].read,
        //       cqe->res, r_data[(int)cqe->user_data].size,
        //       cqe->res - r_data[(int)cqe->user_data].size,
        //       offset, to_write);
        check_exit(cqe->res, "Error in async operation: %s\n");
        int id = (int)cqe->user_data;
        io_uring_cqe_seen(&ring, cqe);
        if (r_data[id].read) {
            if (submit_write(iov, r_data, id, output, &ring) < 0) {
                return -1;
            }
            io_uring_submit(&ring);
        }
        else {
            to_write -= cqe->res;
            if (offset != file_size) {
                if (submit_read(read_size, file_size, &offset, iov, r_data, id,
                                input, &ring) < 0) {
                    return -1;
                }
                io_uring_submit(&ring);
            }
        }
    }

    free(buffers);
    io_uring_queue_exit(&ring);
    res = close(input);
    check_exit(res, "Error while closing file: %s\n");
    res = close(output);
    check_exit(res, "Error while closing file: %s\n");
    return 0;
}
