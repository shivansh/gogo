# Test to find sum of all even numbers less than n

	.data
nStr:		.asciiz "Enter n: "
n:		.word	0
i:		.word	0
k:		.word	0
l:		.word	0
str:		.asciiz "Sum of all even numbers less than n: "

	.text


	.globl main
	.ent main
main:
	li	$2, 4
	la	$4, nStr
	syscall
	li	$2, 5
	syscall
	move	$3, $2
	sw	$3, n		# spilled n, freed $3
	li	$3, 0		# i -> $3
	sw	$3, i		# spilled i, freed $3
	li	$3, 0		# k -> $3
	# Store dirty variables back into memory
	sw	$3, k

loop:
	lw	$3, i
	lw	$5, n
	bge	$3, $5, exit
	# Store dirty variables back into memory

	lw	$3, i
	rem	$5, $3, 2	# l -> $5
	# Store dirty variables back into memory
	sw	$5, l
	beq	$5, 1, skip

	lw	$3, k
	lw	$5, i
	add	$3, $3, $5	# k -> $3
	addi	$5, $5, 1	# i -> $5
	# Store dirty variables back into memory
	sw	$3, k
	sw	$5, i
	j	loop

skip:
	lw	$3, i
	addi	$3, $3, 1	# i -> $3
	# Store dirty variables back into memory
	sw	$3, i
	j	loop

exit:
	li	$2, 4
	la	$4, str
	syscall
	li	$2, 1
	lw	$3, k
	move	$4, $3
	syscall
	# Store dirty variables back into memory
	li	$2, 10
	syscall
	.end main
