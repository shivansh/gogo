	.data
a.0:		.word	0
newline.1:	.asciiz "\n"
i.2:		.word	0
t0:		.word	0

	.text

	.globl main
	.ent main
main:
	li	$3, 0		# a.0 -> $3
	sw	$3, a.0	# spilled a.0, freed $3
	li	$3, 0		# i.2 -> $3
	# Store dirty variables back into memory
	sw	$3, i.2

l2:
	lw	$3, i.2
	bge	$3, 4, l0
	# Store dirty variables back into memory

	li	$3, 1		# t0 -> $3
	# Store dirty variables back into memory
	sw	$3, t0
	j	l1

l0:
	li	$3, 0		# t0 -> $3
	# Store dirty variables back into memory
	sw	$3, t0

l1:
	lw	$3, t0
	blt	$3, 1, l3
	# Store dirty variables back into memory

	lw	$3, a.0
	addi	$3, $3, 1	# a.0 -> $3
	li	$2, 1
	move	$4, $3
	syscall
	li	$2, 4
	la	$4, newline.1
	syscall
	lw	$5, i.2
	addi	$5, $5, 1	# i.2 -> $5
	# Store dirty variables back into memory
	sw	$3, a.0
	sw	$5, i.2
	j	l2

l3:
	li	$2, 10
	syscall
	.end main
