	.data
a.0:		.space	8
t0:		.word	0
t1:		.word	0
t2:		.word	0
t3:		.word	0

	.text

	.globl main
	.ent main
main:
	la	$3, a.0
	lw	$5, 0($3)	# variable <- array
	li	$5, 1		# t0 -> $5
	sw	$5, 0($3)	# variable -> array
	# Store dirty variables back into memory
	sw	$5, t0

l4:
	la	$3, a.0
	lw	$5, 0($3)	# variable <- array
	# Store dirty variables back into memory
	sw	$5, t1
	bne	$5, 1, l0

	li	$3, 1		# t2 -> $3
	# Store dirty variables back into memory
	sw	$3, t2
	j	l1

l0:
	li	$3, 0		# t2 -> $3
	# Store dirty variables back into memory
	sw	$3, t2

l1:
	lw	$3, t2
	blt	$3, 1, l4
	# Store dirty variables back into memory

	j	l5

l5:
	la	$3, a.0
	lw	$5, 0($3)	# variable <- array
	li	$2, 1
	move	$4, $5
	syscall
	# Store dirty variables back into memory
	sw	$5, t3
	li	$2, 10
	syscall
	.end main
