	.data
a.0:		.word	0
b.1:		.word	0
newline.2:	.asciiz "\n"

	.text

	.globl main
	.ent main
main:
	li	$3, 1		# a.0 -> $3
	sw	$3, a.0	# spilled a.0, freed $3
	li	$3, 2		# b.1 -> $3
	# Store dirty variables back into memory
	sw	$3, b.1
	j	l2

l1:
	li	$2, 1
	lw	$3, a.0
	move	$4, $3
	syscall
	li	$2, 4
	la	$4, newline.2
	syscall
	# Store dirty variables back into memory
	li	$2, 10
	syscall
	.end main
l2:
	li	$3, 4		# a.0 -> $3
	li	$2, 1
	sw	$3, a.0	# spilled a.0, freed $3
	lw	$3, b.1
	move	$4, $3
	syscall
	li	$2, 4
	la	$4, newline.2
	syscall
	# Store dirty variables back into memory
	j	l1


