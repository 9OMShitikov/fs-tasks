#include <stdio.h>
#include <fcntl.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>
#include <inttypes.h>
#include <sys/stat.h>
#include <liburing.h>
#include <errno.h>

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
    if (sqe == NULL) {
        fprintf(stderr, "could not get SQE.\n");
        return -1;
    }

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
    if (sqe == NULL) {
        fprintf(stderr, "could not get SQE.\n");
        return -1;
    }

    r_data[id].read = 0;
    io_uring_prep_write_fixed(sqe, fd, iov[id].iov_base, r_data[id].size, r_data[id].offset, id);
    sqe->user_data = (uint64_t) id;
    return 0;
}

int main(int argc, char* argv[]) {
    int N = 4;
    size_t read_size = 256 * 1024;

    if (argc != 3 && argc != 5) {
        fprintf(stderr, "Arguments were not provided.\n"
            "Arguments: in - name of input file, out - name of output file\n"
            "Optional arguments: N - number of buffers, read_size - size of buffer (should be used together)\n");
        return -1;
    }
    if (argc == 5) {
        N = strtol(argv[3], NULL, 10);
        read_size = strtol(argv[4], NULL, 10);
    }

    int input = open(argv[1], O_RDONLY);
    if (input < 0) {
        perror("error while opening input file ");
        return -1;
    }

    int output = open(argv[2], O_WRONLY | O_CREAT | O_TRUNC, 0755);
    if (output < 0) {
        perror("error while opening output file ");
        return -1;
    }

    struct io_uring ring;
    int res = io_uring_queue_init(2 * N, &ring, 0);
    if (res < 0) {
        fprintf(stderr, "error while queue_init: %s\n", strerror(-res));
        return -1;
    }

    char* buffers = malloc(N * read_size);
    if (buffers == NULL) {
        fprintf(stderr, "error while allocating memory for buffer\n");
        return -1;
    }

    struct iovec iov[N];
    for (int i = 0; i < N; ++i) {
        iov[i].iov_len = read_size;
        iov[i].iov_base = buffers + i * read_size;
    }
    res = io_uring_register_buffers(&ring, iov, N);
    if (res < 0) {
        fprintf(stderr, "error registering buffers: %s\n", strerror(-res));
        return -1;
    }

    struct stat st;
    res = fstat(input, &st);
    if (res < 0) {
        perror("error while checking input file size ");
        return -1;
    }
    if (!S_ISREG(st.st_mode)) {
        fprintf(stderr, "wrong input file type. Input file should be regular\n");
        return -1;
    }

    size_t offset = 0;
    size_t to_write = st.st_size;
    size_t file_size = st.st_size;

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

        if (res < 0) {
            fprintf(stderr, "error waiting for completion: %s\n", strerror(-res));
            return -1;
        }

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
    if (res == -1) {
        perror("error while closing input file ");
        return -1;
    }
    res = close(output);
    if (res == -1) {
        perror("error while closing output file ");
        return -1;
    }
    return 0;
}
